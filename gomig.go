package main

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/multierr"
	"gomig/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Gomig struct {
	ctx       context.Context
	conn      *sql.DB
	dbName    string
	tableName string

	err error
}

type File struct {
	path  string
	name  string
	mType Action
}

type Files []*File

func (f Files) Sort() {
	sort.Slice(f, func(i, j int) bool {
		return f[i].name < f[j].name
	})
}

func NewGomig(ctx context.Context, db *database, dbName string) *Gomig {
	conn, err := db.Connect()
	if err != nil {
		return &Gomig{err: err}
	}
	g := &Gomig{ctx: ctx, conn: conn, dbName: dbName, tableName: db.TableName}
	if err := g.checkMigrationsTable(); err != nil {
		return &Gomig{err: err}
	}
	rows, err := conn.QueryContext(ctx, fmt.Sprintf(querySelectMigrationValues, dbName, db.TableName))
	if err != nil {
		return &Gomig{err: err}
	}
	defer util.Close(rows)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return &Gomig{err: rows.Err()}
		}
		dbMigrationFiles = append(dbMigrationFiles, value)
	}
	return g
}

func walk(migrationType Action, files *Files) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		chunks := strings.Split(info.Name(), ".")
		fName, mType := chunks[0], chunks[1]
		if migrationType.is(mType) && UP.is(mType) {
			for _, fileName := range dbMigrationFiles {
				if fileName == fName {
					return nil
				}
			}
			*files = append(*files, &File{path: path, name: fName, mType: migrationType})
		} else if migrationType.is(mType) && DOWN.is(mType) {
			for _, fileName := range dbMigrationFiles {
				if fileName == fName {
					*files = append(*files, &File{path: path, name: fName, mType: migrationType})
					continue
				}
			}
		}
		return nil
	}
}

func (g *Gomig) exec(migrationType Action) error {
	if g.err != nil {
		return g.err
	}
	var mFiles Files
	if err := filepath.Walk(fmt.Sprintf("%s/%s", migrationDir, g.dbName), walk(migrationType, &mFiles)); err != nil {
		return err
	}
	if len(mFiles) == 0 {
		return ErrNothing
	}
	mFiles.Sort()
	for _, file := range mFiles {
		if err := g.handle(file); err != nil {
			return err
		}
		if file.mType == UP {
			g.err = multierr.Append(g.err, g.Migration(file.name, applyMigration))
		} else if file.mType == DOWN {
			g.err = multierr.Append(g.err, g.Migration(file.name, excludeMigration))
		}
	}
	return g.err
}

func (g *Gomig) handle(file *File) error {
	b, err := ioutil.ReadFile(file.path)
	if err != nil {
		return err
	}
	_, err = g.conn.ExecContext(g.ctx, string(b))
	return err
}

func (g *Gomig) checkMigrationsTable() error {
	var result string
	queryShow := fmt.Sprintf("SHOW TABLES LIKE \"%s\"", g.tableName)
	err := g.conn.QueryRowContext(g.ctx, queryShow).Scan(&result)
	if err == nil {
		return nil
	} else if err != sql.ErrNoRows {
		return err
	}

	queryCreate := fmt.Sprintf(queryCreateTable, g.dbName, g.tableName)
	_, err = g.conn.ExecContext(g.ctx, queryCreate)
	return err
}

func (g *Gomig) Migration(filename string, cb migrationCallback) error {
	return cb(g.ctx, g.conn, g.dbName, g.tableName, filename)
}
