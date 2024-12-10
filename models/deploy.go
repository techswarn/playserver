package models

import (
	"time"
)

//Basic struct for deployment queue implementation
type Deploy struct {
	Id string `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Image string `json:"image"`
	Namespace string `json:namespace`
	CreatedAt time.Time `json:"createdat"`
}