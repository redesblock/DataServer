package models

import (
	"gorm.io/gorm"
	"time"
)

type Bucket struct {
	ID        uint `json:"id" gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string         `json:"name"`
	Access    bool           `json:"access"`
	Network   string         `json:"network"`
	Area      string         `json:"area"`

	UserID uint `json:"-"`

	TotalSize    uint64 `json:"total_size" gorm:"-"`
	TotalNum     uint64 `json:"total_num" gorm:"-"`
	TotalSizeStr string `json:"total_size_str" gorm:"-"`
	Created      string `json:"created_at" gorm:"-"`
	Updated      string `json:"updated_at" gorm:"-"`
}

func (u *Bucket) AfterFind(tx *gorm.DB) (err error) {
	u.Created = u.CreatedAt.Format(TIME_FORMAT)
	u.Updated = u.UpdatedAt.Format(TIME_FORMAT)

	type Result struct {
		Count uint64
		Total uint64
	}
	var ret Result
	if err := tx.Model(&BucketObject{}).Where("status > ?", STATUS_WAIT).Where("bucket_id = ?", u.ID).Where("parent_id = ?", 0).Select("COALESCE(SUM(size),0) AS total, COUNT(id) AS count").Scan(&ret).Error; err != nil {
		return err
	}
	u.TotalSize = ret.Total
	u.TotalNum = ret.Count
	u.TotalSizeStr = ByteSize(u.TotalSize)
	return
}

func FindBuckets(tx *gorm.DB, userID uint, offset int64, limit int64) (total int64, items []*Bucket, err error) {
	err = tx.Model(&Bucket{}).Where("user_id = ?", userID).Order("id DESC").Count(&total).Offset(int(offset)).Limit(int(limit)).Find(&items).Error
	return
}

func FindBucket(tx *gorm.DB, userID uint, id uint) (item *Bucket, err error) {
	ret := tx.Model(&Bucket{}).Where("user_id = ?", userID).Where("id = ?", id).Find(&item)
	if ret.Error != nil {
		err = ret.Error
	}
	if ret.RowsAffected == 0 {
		item = nil
	}
	return
}
