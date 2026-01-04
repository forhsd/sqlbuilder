package repository

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/forhsd/sqlbuilder/connect"
)

type BitBool bool

func InitMysql() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.username"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.dbname"),
	)

	db, err := gorm.Open(mysql.Open(dsn), connect.GormConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

// gorm读数据时调用: 处理 byte(1)类型
func (b BitBool) Value() (driver.Value, error) {
	if b {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

// gorm读数据时调用: 处理 byte(1)类型
func (b *BitBool) Scan(src any) error {
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte assertion")
	}
	*b = v[0] == 1
	return nil
}
