package entity

import "time"

type Team struct {
	ID        string    `json:"-" db:"id"`
	Name      string    `json:"team_name" db:"name"`
	CreatedAt time.Time `json:"-" db:"created_at"`
}
