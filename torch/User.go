package torch

import (

)

type User struct {
	Id             uint64 `json:"id"`
	Username       string `json:"username"`
	password       string
	Access         uint64 `json:"access"`
	SetOnNextLogin bool   `json:"set_on_next_login"`
}
