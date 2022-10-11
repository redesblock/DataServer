package dataservice

type UsedStorage struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"num"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`
}

type UsedTraffic struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"num"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`
}
