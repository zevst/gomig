package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/zevst/zlog"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"io/ioutil"
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
		return nil, nil
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
		return nil, nil
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
	for _, v := range out {
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
	} else if len(files) == 0 {
		zlog.Info("No migrations to apply")
		return nil
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
	} else if len(files) == 0 {
		zlog.Info("No migrations to apply")
		return nil
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

func execMigrations(ctx context.Context, conn *gorm.DB, action Action, files []string) error {
	for _, fp := range files {
		if err := applyMigration(ctx, conn, action, fp); err != nil {
			return err
		}
	}
	return nil
}

func applyMigration(ctx context.Context, conn *gorm.DB, action Action, fp string) error {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	_, err = conn.DB().ExecContext(ctx, string(b))
	if err != nil {
		zlog.Error(action.String(), zap.String("filename", fp), zap.Error(err))
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
		zlog.Info("Migration applied successfully", zap.String("filename", filename))
		if err := tx.Create(&Entity{Value: filename}).Error; err != nil {
			err = multierr.Append(err, tx.Rollback().Error)
			return err
		}
	case DOWN:
		_, fn := filepath.Split(fp)
		filename := strings.TrimSuffix(fn, downSuffix)
		zlog.Info("Migration applied successfully", zap.String("filename", filename))
		where := map[string]interface{}{"value": filename}
		if err := tx.Delete(&Entity{}, where).Error; err != nil {
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
