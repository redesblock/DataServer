package dataservice

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

const TIME_FORMAT = "2006-01-02 15:04:05"

type DataService struct {
	*gorm.DB
}

func New(mode string, dsn string, opts ...gorm.Option) *DataService {
	var conn gorm.Dialector
	switch mode {
	case "sqlite":
		os.MkdirAll(filepath.Dir(dsn), 0755)
		conn = sqlite.Open(dsn)
	case "mysql":
		conn = mysql.Open(dsn)
	case "postgres":
		conn = postgres.Open(dsn)
	default:
		log.Fatal("Failed to connect to database: ", "invalid db engine. supported types: sqlite, mysql, postgres")
	}

	//opts = append(opts, &gorm.Config{
	//	Logger: logger.Default.LogMode(logger.Error),
	//})

	db, err := gorm.Open(conn, opts...)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&UserAction{})
	db.AutoMigrate(&BillStorage{})
	db.AutoMigrate(&BillTraffic{})
	db.AutoMigrate(&Bucket{})
	db.AutoMigrate(&BucketObject{})
	db.AutoMigrate(&UsedTraffic{})
	db.AutoMigrate(&UsedStorage{})
	db.AutoMigrate(&Voucher{})
	db.AutoMigrate(&ReportTraffic{})

	if err := db.Save(&Voucher{
		ID:      1,
		Node:    "182.140.245.81",
		Voucher: "6a37426428da189639e73cd88012fcbc21211d50a0949a188577db303c73dea0",
		Area:    "China",
		Usable:  true,
	}).Error; err != nil {
		log.Fatal("Failed to init database: ", err)
	}

	if err := db.Save(&User{
		ID:       100,
		Email:    "admin@redeslab.io",
		Password: "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9",
	}).Error; err != nil {
		log.Fatal("Failed to init database: ", err)
	}

	return &DataService{
		DB: db,
	}
}
