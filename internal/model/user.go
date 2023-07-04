package model

import "time"

type User struct {
	ID       int       `json:"-"`
	Name     string    `json:"name"`
	Password string    `json:"password"`
	Created  time.Time `json:"-"`
	Updated  time.Time `json:"-"`
}

type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AccessUuid   string `json:"access_uuid"`
	RefreshUuid  string `json:"refresh_uuid"`
	AtExpires    int64  `json:"at_expires"`
	RtExpires    int64  `json:"rt_expires"`
}

type AccessDetails struct {
	AccessUuid string `json:"access_uuid"`
	UserId     int    `json:"user_id"`
}

type Token struct {
	ID    string `json:"-"`
	Token string `json:"token"`
}
