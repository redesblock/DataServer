package dataservice

type ReportTraffic struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Address    string `json:"address"`
	IP         string `json:"ip"`
	Key        string `json:"key"`
	Uploaded   int64  `json:"node"`
	Downloaded int64  `json:"area"`
	Timestamp  int64  `json:"timestamp"`
}
