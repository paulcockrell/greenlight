package data

import "time"

type Movie struct {
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"titles"`
	Genres    []string  `json:"genres,omitempty"`
	ID        int64     `json:"id"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime"`
	Version   int32     `json:"version"`
}
