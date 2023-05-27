package models

import (
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

const TIME_FORMAT = "2006-01-02 15:04:05"

var db *gorm.DB

func New(mode string, dsn string, opts ...gorm.Option) *gorm.DB {
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

	var err error
	db, err = gorm.Open(conn, opts...)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	if err := Init(db); err != nil {
		log.Fatal("Failed to init database: ", err)
	}

	return db
}

func Init(db *gorm.DB) error {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&UserAction{})
	db.AutoMigrate(&Bucket{})
	db.AutoMigrate(&BucketObject{})
	db.AutoMigrate(&UsedTraffic{})
	db.AutoMigrate(&UsedStorage{})
	db.AutoMigrate(&ReportTraffic{})
	db.AutoMigrate(&Node{})
	db.AutoMigrate(&SignIn{})
	db.AutoMigrate(&Currency{})
	db.AutoMigrate(&Product{})
	db.AutoMigrate(&SpecialProduct{})
	db.AutoMigrate(&Coupon{})
	db.AutoMigrate(&UserCoupon{})
	db.AutoMigrate(&Order{})

	signIns := []*SignIn{
		{
			ID:       1,
			PType:    ProductType_Storage,
			Quantity: 1024 * 1024 * 100,
			Period:   SignInPeriod_Day,
			Enable:   true,
		},
		{
			ID:       2,
			PType:    ProductType_Traffic,
			Quantity: 1024 * 1024 * 100,
			Period:   SignInPeriod_Day,
			Enable:   true,
		},
	}
	currencies := []*Currency{
		{
			ID:      1,
			Symbol:  "USDT",
			Rate:    decimal.NewFromFloat(1),
			Base:    true,
			Payment: PaymentChannel_Crypto,
		},
		{
			ID:      2,
			Symbol:  "MOP",
			Rate:    decimal.NewFromFloat(1000),
			Base:    false,
			Payment: PaymentChannel_Crypto,
		},
		{
			ID:      3,
			Symbol:  "CNY",
			Rate:    decimal.NewFromFloat(7),
			Base:    false,
			Payment: PaymentChannel_Alipay | PaymentChannel_WeChat,
		},
		{
			ID:      4,
			Symbol:  "USD",
			Rate:    decimal.NewFromFloat(1),
			Base:    false,
			Payment: PaymentChannel_Alipay | PaymentChannel_WeChat,
		},
	}
	products := []*Product{
		{
			ID:         1,
			PType:      ProductType_Storage,
			Quantity:   1024 * 1024,
			Price:      decimal.NewFromFloat(1),
			CurrencyID: 1,
		},
		{
			ID:         2,
			PType:      ProductType_Traffic,
			Quantity:   1024 * 1024,
			Price:      decimal.NewFromFloat(1),
			CurrencyID: 1,
		},
	}
	nodes := []*Node{
		{
			ID:        1,
			Name:      "重庆",
			IP:        "182.140.245.81",
			Port:      1683,
			VoucherID: "6a37426428da189639e73cd88012fcbc21211d50a0949a188577db303c73dea0",
			Zone:      "China",
			Usable:    true,
		},
	}
	users := []*User{
		{
			ID:       1,
			Email:    "admin@redeslab.io",
			Password: "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9",
			Role:     UserRole_Admin,
		},
	}

	for _, item := range signIns {
		var count int64
		if err := db.Model(&SignIn{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Save(&item).Error; err != nil {
				return err
			}
		}
	}

	for _, item := range currencies {
		var count int64
		if err := db.Model(&Currency{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Save(&item).Error; err != nil {
				return err
			}
		}
	}

	for _, item := range products {
		var count int64
		if err := db.Model(&Product{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Save(&item).Error; err != nil {
				return err
			}
		}
	}

	for _, item := range nodes {
		var count int64
		if err := db.Model(&Node{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Save(&item).Error; err != nil {
				return err
			}
		}
	}

	for _, item := range users {
		var count int64
		if err := db.Model(&User{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Save(&item).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
