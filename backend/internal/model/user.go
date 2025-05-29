package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	UserId     uint32    `json:"user_id"`
	Phone      string    `json:"phone"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (u *User) BeforeCreate(db *gorm.DB) error {
	u.CreateTime = time.Now()
	u.UpdateTime = time.Now()
	return nil
}

func (u *User) BeforeUpdate(db *gorm.DB) error {
	u.UpdateTime = time.Now()
	return nil
}
