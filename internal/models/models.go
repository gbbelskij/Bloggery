package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	UserId  uuid.UUID `json:"user_id" env-required:"true"`
	Title   string    `json:"title" env=required:"true"`
	Content string    `json:"content"`
}

type DbPost struct {
	PostId    uuid.UUID `json:"post_id" env-required:"true"`
	UserId    uuid.UUID `json:"user_id" env-required:"true"`
	Title     string    `json:"title" env-required:"true"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
