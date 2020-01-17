package driver

import "github.com/jinzhu/gorm"
import _ "github.com/jinzhu/gorm/dialects/postgres"

const PgSQL = "postgres"

func PgSQLConn(args ...interface{}) (db *gorm.DB, err error) {
	return gorm.Open(PgSQL, args...)
}
