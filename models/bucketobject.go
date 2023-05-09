package models

import (
	"github.com/spf13/viper"

	"gorm.io/gorm"
)

const (
	STATUS_UNKOWN int = iota
	STATUS_WAIT
	STATUS_UPLOAD
	STATUS_UPLOADED
	STATUS_PIN
	STATUS_PINED
	STATUS_FAIL_PINED
)

var Statuses = []string{
	"Unkown",
	"Wait",
	"Uploading",
	"Pinning",
	"Pined",
	"Pined",
	"Unpin",
}

type BucketObject struct {
	gorm.Model
	Name     string `json:"name"`
	CID      string `json:"cid"`
	Size     uint64 `json:"size"`
	Status   int    `json:"-"`
	AssetID  string `json:"asset_id"`
	ParentID uint   `json:"-"`
	BucketID uint   `json:"-"`

	Created        string    `json:"created_at" gorm:"-"`
	Updated        string    `json:"updated_at" gorm:"-"`
	TotalSize      uint64    `json:"total_size" gorm:"-"`
	TotalNum       uint64    `json:"total_num" gorm:"-"`
	TotalSizeStr   string    `json:"total_size_str" gorm:"-"`
	IsFolder       bool      `json:"is_folder" gorm:"-"`
	URL            string    `json:"url" gorm:"-"`
	SizeStr        string    `json:"size_str" gorm:"-"`
	StatusStr      string    `json:"status" gorm:"-"`
	Parents        []*Parent `json:"level" gorm:"-"`
	UplinkProgress int       `json:"uplinkProgress"`
}

type Parent struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (u *BucketObject) AfterFind(tx *gorm.DB) (err error) {
	u.Created = u.CreatedAt.Format(TIME_FORMAT)
	u.Updated = u.UpdatedAt.Format(TIME_FORMAT)
	if len(u.AssetID) == 0 && len(u.CID) == 0 {
		u.IsFolder = true
	} else {
		u.StatusStr = Statuses[u.Status]
	}
	u.SizeStr = ByteSize(u.Size)
	if len(u.CID) > 0 {
		u.URL = viper.GetString("gateway") + "mop/" + u.CID + "/"
	}
	return
}

func FindBucketObjects(tx *gorm.DB, bucketID uint, fid uint, offset int64, limit int64) (total int64, items []*BucketObject, err error) {
	err = tx.Model(&BucketObject{}).Where("bucket_id = ?", bucketID).Where("parent_id = ?", fid).Order("id DESC").Count(&total).Offset(int(offset)).Limit(int(limit)).Find(&items).Error
	return
}

func FindBucketObject(tx *gorm.DB, bucketID uint, fid uint) (item *BucketObject, err error) {
	ret := tx.Model(&BucketObject{}).Where("bucket_id = ?", bucketID).Where("id = ?", fid).Find(&item)
	if ret.Error != nil {
		err = ret.Error
	}
	if ret.RowsAffected == 0 {
		item = nil
		return
	}

	parents := []*Parent{
		{ID: item.ID, Name: item.Name},
	}
	if item.IsFolder {
		type Result struct {
			Count uint64
			Total uint64
		}

		var rt Result
		tx.Model(&BucketObject{}).Where("bucket_id = ?", bucketID).Where("parent_id = ?", fid).Where("status > ?", STATUS_WAIT).Select("COALESCE(SUM(size),0) AS total, COUNT(id) AS count").Scan(&rt)
		item.TotalNum = rt.Count
		item.TotalSize = rt.Total
		item.TotalSizeStr = ByteSize(item.TotalSize)

		parentID := item.ParentID
		for parentID > 0 {
			var t BucketObject
			if err = tx.Model(&BucketObject{}).Where("bucket_id = ?", bucketID).Where("id = ?", parentID).Find(&t).Error; err != nil {
				item = nil
				return
			}
			parents = append(parents, &Parent{ID: t.ID, Name: t.Name})
			parentID = t.ParentID
		}
	}

	var t Bucket
	if err = tx.Model(&Bucket{}).Where("id = ?", item.BucketID).Find(&t).Error; err != nil {
		item = nil
		return
	}
	parents = append(parents, &Parent{ID: t.ID, Name: t.Name})

	cnt := len(parents)
	item.Parents = make([]*Parent, cnt)
	for i := cnt; i > 0; i-- {
		item.Parents[cnt-i] = parents[i-1]
	}
	return
}
