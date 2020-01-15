package main

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/multierr"
)

const queryCreateTable = "CREATE TABLE `%s`.`%s` (\n" +
	"`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,\n" +
	"`applied_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n" +
	"`value` varchar(255) NOT NULL,\n" +
	"PRIMARY KEY (`id`),\n" +
	"UNIQUE KEY `uix_value` (`value`) USING BTREE)\n" +
	"ENGINE=InnoDB\n" +
	"DEFAULT CHARSET=utf8mb4\n" +
	"COLLATE=utf8mb4_general_ci;"

const querySelectMigrationValues = "SELECT value FROM `%s`.`%s`"
const queryInsertMigrationValues = "INSERT INTO `%s`.`%s` (value) VALUES(?)"
const queryDeleteMigrationValues = "DELETE FROM `%s`.`%s` WHERE value=?"

type migrationCallback func(ctx context.Context, db *sql.DB, dbName, tableName, fileName string) error

func applyMigration(ctx context.Context, db *sql.DB, dbName, tableName, fileName string) error {
	apply := fmt.Sprintf(queryInsertMigrationValues, dbName, tableName)
	return migration(ctx, db, apply, fileName)
}

func excludeMigration(ctx context.Context, db *sql.DB, dbName, tableName, fileName string) error {
	apply := fmt.Sprintf(queryDeleteMigrationValues, dbName, tableName)
	return migration(ctx, db, apply, fileName)
}

func migration(ctx context.Context, db *sql.DB, query, fileName string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	_, err = tx.StmtContext(ctx, stmt).ExecContext(ctx, fileName)
	if err != nil {
		err = multierr.Append(err, tx.Rollback())
		return err
	}
	return tx.Commit()
}
