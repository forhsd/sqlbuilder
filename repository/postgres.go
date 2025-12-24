package repository

import (
	"fmt"
	"time"

	"github.com/forhsd/sqlbuilder/connect"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPostgres() (db *gorm.DB) {
	pgDsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		viper.GetString("pg.host"),
		viper.GetString("pg.username"),
		viper.GetString("pg.password"),
		viper.GetString("pg.dbname"),
		viper.GetString("pg.port"),
	)

	db, err := gorm.Open(postgres.Open(pgDsn), connect.GormConfig())

	if err != nil {
		panic("failed to connect postgres")
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
