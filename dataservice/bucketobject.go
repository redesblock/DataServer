package dataservice

import (
	"os"
	"time"

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
	"Pending",
	"Pined",
	"Pined",
	"Unpin",
}

type BucketObject struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CID       string    `json:"cid"`
	Size      uint64    `json:"size"`
	Status    int       `json:"-"`
	Chunks    int       `json:"chunks"`
	AssetID   string    `json:"asset_id"`
	ParentID  uint      `json:"-"`
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
	BucketID  uint      `json:"-"`
	Traffic   uint64    `json:"traffic"`

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
	if len(u.AssetID) == 0 && len(u.CID) == 0 {
		u.IsFolder = true
	} else {
		u.StatusStr = Statuses[u.Status]
	}
	u.Created = u.CreatedAt.Format(TIME_FORMAT)
	u.Updated = u.UpdatedAt.Format(TIME_FORMAT)
	u.SizeStr = ByteSize(u.Size)
	if len(u.CID) > 0 {
		gateway, ok := os.LookupEnv("DATA_SERVER_GATEWAY")
		if !ok {
			gateway = "https://gateway.mopweb3.com:13443/"
		}
		u.URL = gateway + "access/" + u.CID + "/"
	}
	return
}

func (s *DataService) FindBucketObjects(bucketID uint, fid uint, offset int64, limit int64) (total int64, items []*BucketObject, err error) {
	err = s.Model(&BucketObject{}).Where("bucket_id = ?", bucketID).Where("parent_id = ?", fid).Order("id DESC").Count(&total).Offset(int(offset)).Limit(int(limit)).Find(&items).Error
	return
}

func (s *DataService) FindBucketObject(bucketID uint, fid uint) (item *BucketObject, err error) {
	ret := s.Model(&BucketObject{}).Where("bucket_id = ?", bucketID).Where("id = ?", fid).Find(&item)
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
		s.Model(&BucketObject{}).Where("bucket_id = ?", bucketID).Where("parent_id = ?", fid).Where("status > ?", STATUS_WAIT).Select("COALESCE(SUM(size),0) AS total, COUNT(id) AS count").Scan(&rt)
		item.TotalNum = rt.Count
		item.TotalSize = rt.Total
		item.TotalSizeStr = ByteSize(item.TotalSize)

		parentID := item.ParentID
		for parentID > 0 {
			var t BucketObject
			if err = s.Model(&BucketObject{}).Where("bucket_id = ?", bucketID).Where("id = ?", parentID).Find(&t).Error; err != nil {
				item = nil
				return
			}
			parents = append(parents, &Parent{ID: t.ID, Name: t.Name})
			parentID = t.ParentID
		}
	}

	var t Bucket
	if err = s.Model(&Bucket{}).Where("id = ?", item.BucketID).Find(&t).Error; err != nil {
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
