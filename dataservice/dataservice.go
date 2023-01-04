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

	if err := db.Save(&Voucher{
		ID:      1,
		Node:    "183.131.181.164",
		Voucher: "15075cfe5d0b95fc200d6dec9b343ddb831d3f1766ba96e381d4c1a3f81fb13a",
		Area:    "China",
		Usable:  true,
	}).Error; err != nil {
		log.Fatal("Failed to init database: ", err)
	}

	return &DataService{
		DB: db,
	}
}
