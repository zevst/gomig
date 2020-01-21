package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"go.uber.org/multierr"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const upSuffix = ".up.sql"
const downSuffix = ".down.sql"

type matchedFiles []string

func (m matchedFiles) Len() int {
	return len(m)
}

func (m matchedFiles) Less(i, j int) bool {
	first := strings.Split(strings.Split(m[i], "_")[0], "/")
	second := strings.Split(strings.Split(m[j], "_")[0], "/")
	return first[len(first)-1] < second[len(second)-1]
}

func (m matchedFiles) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func getUpFiles(conn *gorm.DB) ([]string, error) {
	pattern := fmt.Sprintf("%s/%s/*%s", migrationDir, conn.Dialect().CurrentDatabase(), upSuffix)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	} else if len(matches) == 0 {
		return nil, ErrFilesNotFound
	}
	var in []string
	files := make(map[string]string)
	for _, fp := range matches {
		_, fn := filepath.Split(fp)
		filename := strings.TrimSuffix(fn, upSuffix)
		in = append(in, filename)
		files[filename] = fp
	}
	var migrations []Entity
	if err := conn.Where("value IN (?)", in).Find(&migrations).Error; err != nil {
		return nil, err
	}
	for _, migration := range migrations {
		delete(files, migration.Value)
	}
	if len(files) == 0 {
		return nil, ErrNothing
	}
	var out []string
	for _, v := range files {
		out = append(out, v)
	}
	sort.Sort(matchedFiles(out))
	return out, nil
}

func getDownFiles(conn *gorm.DB) ([]string, error) {
	var migrations []Entity
	if err := conn.Find(&migrations).Error; err != nil {
		return nil, err
	}
	pattern := fmt.Sprintf("%s/%s/*%s", migrationDir, conn.Dialect().CurrentDatabase(), downSuffix)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	} else if len(matches) == 0 {
		return nil, ErrFilesNotFound
	} else if len(migrations) == 0 {
		return nil, ErrNothing
	}
	files := make(map[string]string)
	for _, fp := range matches {
		_, fn := filepath.Split(fp)
		filename := strings.TrimSuffix(fn, downSuffix)
		files[filename] = fp
	}
	out := make(map[string]string)
	for _, migration := range migrations {
		fp, ok := files[migration.Value]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("cannot find file: %s%s", migration.Value, downSuffix))
		}
		out[migration.Value] = fp
	}

	var res []string
	for _, v := range files {
		res = append(res, v)
	}
	sort.Sort(sort.Reverse(matchedFiles(res)))
	return res, err
}

func up(ctx context.Context, db *database) error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}
	files, err := getUpFiles(conn)
	if err != nil {
		return err
	}
	return execMigrations(ctx, conn, UP, files)
}

func down(ctx context.Context, db *database) error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}
	files, err := getDownFiles(conn)
	if err != nil {
		return err
	}
	return execMigrations(ctx, conn, DOWN, files)
}

func apply(ctx context.Context, db *database, file string) error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}
	_, fn := filepath.Split(file)
	var action Action
	if strings.HasSuffix(fn, upSuffix) {
		action = UP
	} else if strings.HasSuffix(fn, downSuffix) {
		action = DOWN
	} else {
		return ErrUndefinedMigrationType
	}

	return applyMigration(ctx, conn, action, file)
}

func execMigrations(ctx context.Context, conn *gorm.DB, action Action, files []string) (err error) {
	for _, fp := range files {
		err = multierr.Append(err, applyMigration(ctx, conn, action, fp))
	}
	return err
}

func applyMigration(ctx context.Context, conn *gorm.DB, action Action, fp string) error {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	_, err = conn.DB().ExecContext(ctx, string(b))
	if err != nil {
		return err
	}
	return updateMigrationList(ctx, conn, action, fp)
}

func updateMigrationList(ctx context.Context, conn *gorm.DB, action Action, fp string) error {
	tx := conn.BeginTx(ctx, nil)
	switch action {
	case UP:
		_, fn := filepath.Split(fp)
		filename := strings.TrimSuffix(fn, upSuffix)
		log.Printf("Migration applied successfully: %s", filename)
		if err := tx.Create(&Entity{Value: filename}).Error; err != nil {
			err = multierr.Append(err, tx.Rollback().Error)
			return err
		}
	case DOWN:
		_, fn := filepath.Split(fp)
		filename := strings.TrimSuffix(fn, downSuffix)
		log.Printf("Migration applied successfully: %s", filename)
		if err := tx.Delete(&Entity{Value: filename}).Error; err != nil {
			err = multierr.Append(err, tx.Rollback().Error)
			return err
		}
	}
	return tx.Commit().Error
}

func create(base, name, out string) error {
	return createMigration(fmt.Sprintf("%s/%s/%d_%s", out, base, time.Now().Unix(), name))
}

func createMigration(name string) error {
	up := fmt.Sprintf("%s%s", name, upSuffix)
	down := fmt.Sprintf("%s%s", name, downSuffix)
	if err := createFile(up); err != nil {
		return err
	}
	return createFile(down)
}

func createFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = file.WriteString("-- ")
	return err
}
