package entity

import "time"

type Team struct {
	Name      string    `json:"team_name" db:"name"`
	CreatedAt time.Time `json:"-" db:"created_at"`
}
