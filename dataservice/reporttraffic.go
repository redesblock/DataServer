package dataservice

type ReportTraffic struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Token      string `json:"token"`
	Uploaded   int64  `json:"node"`
	Downloaded int64  `json:"area"`
	Timestamp  int64  `json:"timestamp"`
}
