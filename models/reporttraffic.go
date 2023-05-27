package models

import (
	"gorm.io/gorm"
	"time"
)

type ReportTraffic struct {
	ID            uint `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	Token         string         `json:"token"`
	Uploaded      int64          `json:"uploaded"`
	UploadedCnt   int64          `json:"uploaded_cnt"`
	Downloaded    int64          `json:"downloaded"`
	DownloadedCnt int64          `json:"downloaded_cnt"`
	Timestamp     int64          `json:"-"`
	NATAddr       string         `json:"nat_addr"`

	UploadedStr   string `json:"uploaded_str" gorm:"-"`
	DownloadedStr string `json:"downloaded_str" gorm:"-"`
	TimestampStr  string `json:"timestamp" gorm:"-"`
}

func (u *ReportTraffic) AfterFind(tx *gorm.DB) (err error) {
	u.TimestampStr = time.Unix(u.Timestamp, 0).Format(TIME_FORMAT)
	u.DownloadedStr = ByteSize(uint64(u.Downloaded))
	u.UploadedStr = ByteSize(uint64(u.Uploaded))
	return
}
