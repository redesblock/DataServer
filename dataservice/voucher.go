package dataservice

type Voucher struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Voucher string `json:"voucher" gorm:"unique"`
	Node    string `json:"node"`
	Area    string `json:"area"`
	Usable  bool   `json:"usable"`
}

func (s *DataService) FindVouchers() (items []*Voucher, err error) {
	err = s.Model(&Voucher{}).Where("usable = true").Order("id DESC").Find(&items).Error
	return
}

func (s *DataService) FindAreas() (items []string, err error) {
	err = s.Model(&Voucher{}).Select("area").Where("usable = true").Order("area DESC").Find(&items).Error
	return
}
