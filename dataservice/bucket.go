package dataservice

import (
	"github.com/dustin/go-humanize"
	"gorm.io/gorm"
	"time"
)

type Bucket struct {
	ID uint `json:"id" gorm:"primaryKey"`
	//Email     string    `json:"email" gorm:"index"`
	Name      string    `json:"name"`
	Access    bool      `json:"access"`
	Network   string    `json:"network"`
	Area      string    `json:"area"`
	CreatedAt time.Time `json:"-"`
	UserID    uint      `json:"-"`

	Created      string `json:"created_at" gorm:"-"`
	TotalSize    uint64 `json:"total_size" gorm:"-"`
	TotalNum     uint64 `json:"total_num" gorm:"-"`
	TotalSizeStr string `json:"total_size_str" gorm:"-"`
}

func (u *Bucket) AfterFind(tx *gorm.DB) (err error) {
	u.Created = u.CreatedAt.Format(TIME_FORMAT)

	type Result struct {
		Count uint64
		Total uint64
	}
	var ret Result
	if err := tx.Model(&BucketObject{}).Where("c_id != ''").Where("bucket_id = ?", u.ID).Where("parent_id = ?", 0).Select("COALESCE(SUM(size),0) AS total, COUNT(id) AS count").Scan(&ret).Error; err != nil {
		return err
	}
	u.TotalSize = ret.Total
	u.TotalNum = ret.Count
	u.TotalSizeStr = humanize.Bytes(u.TotalSize)
	return
}

func (s *DataService) FindBuckets(userID uint, offset int64, limit int64) (total int64, items []*Bucket, err error) {
	err = s.Model(&Bucket{}).Where("user_id = ?", userID).Order("id DESC").Count(&total).Offset(int(offset)).Limit(int(limit)).Find(&items).Error
	return
}

func (s *DataService) FindBucket(userID uint, id uint) (item *Bucket, err error) {
	ret := s.Model(&Bucket{}).Where("user_id = ?", userID).Where("id = ?", id).Find(&item)
	if ret.Error != nil {
		err = ret.Error
	}
	if ret.RowsAffected == 0 {
		item = nil
	}
	return
}
