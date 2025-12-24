package connect

import (
	"gorm.io/gorm/logger"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func GormConfig() (config *gorm.Config) {
	return &gorm.Config{
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		DryRun:                                   false,
		NamingStrategy: schema.NamingStrategy{
			//TablePrefix: "t_",   // table name prefix, table for `User` would be `t_users`
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
			//NoLowerCase:   true,                              // skip the snake_casing of names
			//NameReplacer:  strings.NewReplacer("CID", "Cid"), // use name replacer to change struct/field name before convert it to db name
		},
		Logger: logger.Default.LogMode(logger.Info),
	}
}

func MysqlConn() (db *gorm.DB) {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	mysqlDsn := "root:mysqlHolder@tcp(192.168.100.100:3388)/holder_bi?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(mysqlDsn), GormConfig())

	if err != nil {
		panic("failed to connect mysql")
	}

	// 连接池
	sqlDB, _ := db.DB()
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

func PostgresConn() (db *gorm.DB) {
	//pgDsn := "host=192.168.99.60 user=postgres password=postgres dbname=test_bi port=5000 sslmode=disable TimeZone=Asia/Shanghai search_path=test_data"
	//db, err := gorm.Open(postgres.Open(pgDsn), GormConfig())

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=192.168.99.60 user=postgres password=postgres dbname=dwd port=5000 sslmode=disable TimeZone=Asia/Shanghai",
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), GormConfig())

	if err != nil {
		panic("failed to connect postgres")
	}

	// 连接池
	sqlDB, _ := db.DB()
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}
