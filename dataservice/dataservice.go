package dataservice

import (
	"gorm.io/gorm"
)

const TIME_FORMAT = "2006-01-02 15:04:05"

type DataService struct {
	*gorm.DB
}

//func New(mode string, dsn string, opts ...gorm.Option) *DataService {
//	var conn gorm.Dialector
//	switch mode {
//	case "sqlite":
//		os.MkdirAll(filepath.Dir(dsn), 0755)
//		conn = sqlite.Open(dsn)
//	case "mysql":
//		conn = mysql.Open(dsn)
//	case "postgres":
//		conn = postgres.Open(dsn)
//	default:
//		log.Fatal("Failed to connect to database: ", "invalid db engine. supported types: sqlite, mysql, postgres")
//	}
//
//	//opts = append(opts, &gorm.Config{
//	//	Logger: logger.Default.LogMode(logger.Error),
//	//})
//
//	db, err := gorm.Open(conn, opts...)
//	if err != nil {
//		log.Fatal("Failed to connect to database: ", err)
//	}
//
//	db.AutoMigrate(&models.User{})
//	db.AutoMigrate(&models.UserAction{})
//	db.AutoMigrate(&models.Bucket{})
//	db.AutoMigrate(&models.BucketObject{})
//	db.AutoMigrate(&models.DailyUsedTraffic{})
//	db.AutoMigrate(&models.DailyUsedStorage{})
//	db.AutoMigrate(&models.ReportTraffic{})
//	db.AutoMigrate(&models.Node{})
//	db.AutoMigrate(&models.SignIn{})
//	db.AutoMigrate(&models.Currency{})
//	db.AutoMigrate(&models.Product{})
//	db.AutoMigrate(&models.SpecialProduct{})
//	db.AutoMigrate(&models.Coupon{})
//	db.AutoMigrate(&models.UserCoupon{})
//	db.Save([]*models.SignIn{
//		&models.SignIn{
//			Model: gorm.Model{
//				ID: 1,
//			},
//			PType:    models.ProductType_Storage,
//			Quantity: 1024 * 1024 * 100,
//			Period:   models.SignInPeriod_Day,
//			Enable:   true,
//		},
//		&models.SignIn{
//			Model: gorm.Model{
//				ID: 2,
//			},
//			PType:    models.ProductType_Traffic,
//			Quantity: 1024 * 1024 * 100,
//			Period:   models.SignInPeriod_Day,
//			Enable:   true,
//		},
//	})
//	db.Save([]*models.Currency{
//		&models.Currency{
//			Model: gorm.Model{
//				ID: 1,
//			},
//			Symbol:  "USDT",
//			Rate:    decimal.NewFromFloat(1),
//			Base:    true,
//			Payment: models.PaymentChannel_Crypto,
//		},
//		&models.Currency{
//			Model: gorm.Model{
//				ID: 2,
//			},
//			Symbol:  "MOP",
//			Rate:    decimal.NewFromFloat(1000),
//			Base:    false,
//			Payment: models.PaymentChannel_Crypto,
//		},
//		&models.Currency{
//			Model: gorm.Model{
//				ID: 3,
//			},
//			Symbol:  "CNY",
//			Rate:    decimal.NewFromFloat(7),
//			Base:    false,
//			Payment: models.PaymentChannel_Alipay | models.PaymentChannel_WeChat,
//		},
//		&models.Currency{
//			Model: gorm.Model{
//				ID: 4,
//			},
//			Symbol:  "USD",
//			Rate:    decimal.NewFromFloat(1),
//			Base:    false,
//			Payment: models.PaymentChannel_Alipay | models.PaymentChannel_WeChat,
//		},
//	})
//	db.Save([]*models.Product{
//		&models.Product{
//			Model: gorm.Model{
//				ID: 1,
//			},
//			PType:      models.ProductType_Storage,
//			Quantity:   1024 * 1024,
//			Price:      decimal.NewFromFloat(1),
//			CurrencyID: 1,
//		},
//		&models.Product{
//			Model: gorm.Model{
//				ID: 2,
//			},
//			PType:      models.ProductType_Traffic,
//			Quantity:   1024 * 1024,
//			Price:      decimal.NewFromFloat(1),
//			CurrencyID: 1,
//		},
//	})
//
//	if err := db.Save(&Voucher{
//		ID:      1,
//		Node:    "182.140.245.81",
//		Voucher: "6a37426428da189639e73cd88012fcbc21211d50a0949a188577db303c73dea0",
//		Area:    "China",
//		Usable:  true,
//	}).Error; err != nil {
//		log.Fatal("Failed to init database: ", err)
//	}
//
//	if err := db.Save(&User{
//		ID:       100,
//		Email:    "admin@redeslab.io",
//		Password: "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9",
//	}).Error; err != nil {
//		log.Fatal("Failed to init database: ", err)
//	}
//
//	return &DataService{
//		DB: db,
//	}
//}
