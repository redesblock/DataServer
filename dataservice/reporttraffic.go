package dataservice

type ReportTraffic struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	Token         string `json:"token"`
	Uploaded      int64  `json:"uploaded"`
	UploadedCnt   int64  `json:"uploaded_cnt"`
	Downloaded    int64  `json:"downloaded"`
	DownloadedCnt int64  `json:"downloaded_cnt"`
	Timestamp     int64  `json:"timestamp"`
	NATAddr       string `json:"nat_addr"`
}
