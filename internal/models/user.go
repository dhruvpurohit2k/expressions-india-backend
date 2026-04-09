package models

type User struct {
	ID       string `json:"id"`
	EmailID  string `json:"emailId"`
	Password string `json:"-"`
	IsAdmin  bool   `gorm:"default:false" json:"isAdmin"`
}
