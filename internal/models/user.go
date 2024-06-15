package models

type User struct {
	ID          int      `json:"id"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Birthday    string   `json:"birthday"`
	Subscribers []string `json:"subscribers"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserSubscribe struct {
	TargetEmail string `json:"email"`
}
