package models

import "time"

type User struct {
	Id   *uint64 `json:"id"`
	Name string  `json:"name"`

	UpdatedAt *time.Time `json:"updated_at"`
	CreatedAt *time.Time `json:"created_at"`
}
