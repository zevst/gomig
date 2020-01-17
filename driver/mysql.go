package driver

import "github.com/jinzhu/gorm"
import _ "github.com/jinzhu/gorm/dialects/mysql"

const MySQL = "mysql"

func MySQLConn(args ...interface{}) (db *gorm.DB, err error) {
	return gorm.Open(MySQL, args...)
}
