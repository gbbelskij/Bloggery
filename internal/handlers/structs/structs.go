package structs

import "time"

type PaginationParams struct {
	Limit   int       `form:"limit" binding:"required,min=1,max=5"`
	Cursor  time.Time `form:"cursor" binding:"required"`
	Reverse bool      `form:"reverse" binding:"omitempty"`
}
