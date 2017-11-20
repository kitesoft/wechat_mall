package main

import (
	"time"
)

// WeAppUser 微信用户
type WeAppUser struct {
	OpenID    string `json:"openId"`
	Nickname  string `json:"nickName"`
	Gender    int    `json:"gender"`
	AvatarURL string `json:"avatarUrl"`
}

// User 用户
type User struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt"`
	ContactID string     `json:"contactId"` //默认地址
	OpenID    string     `json:"openId"`
	Nickname  string     `json:"nickname"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Password  string     `json:"password"`
	Token     string     `json:"token"`
	Sex       bool       `json:"sex"`
	Subscribe bool       `json:"subscribe"`
	Status    int        `json:"status"`
	Lastip    string     `json:"lastip"`
}
