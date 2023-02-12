package dataservice

import (
	"gorm.io/gorm"
	"time"
)

type ReportTraffic struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	Token         string `json:"token"`
	Uploaded      int64  `json:"-"`
	UploadedCnt   int64  `json:"uploaded_cnt"`
	Downloaded    int64  `json:"-"`
	DownloadedCnt int64  `json:"downloaded_cnt"`
	Timestamp     int64  `json:"timestamp"`
	NATAddr       string `json:"nat_addr"`

	UploadedStr   string `json:"uploaded" gorm:"-"`
	DownloadedStr string `json:"downloaded" gorm:"-"`
	TimestampStr  string `json:"timestamp_str" gorm:"-"`
}

func (u *ReportTraffic) AfterFind(tx *gorm.DB) (err error) {
	u.TimestampStr = time.Unix(u.Timestamp, 0).Format(TIME_FORMAT)
	u.DownloadedStr = ByteSize(uint64(u.Downloaded))
	u.UploadedStr = ByteSize(uint64(u.Uploaded))
	return
}

func (s *DataService) FindTraffics(offset int64, limit int64) (total int64, items []*ReportTraffic, err error) {
	err = s.Model(&ReportTraffic{}).Order("timestamp DESC, nat_addr").Count(&total).Offset(int(offset)).Limit(int(limit)).Find(&items).Error
	return
}
