package models

import "time"

type Objects struct {
	UserId          string
	ObjectName      string
	DownloadedCount int
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
