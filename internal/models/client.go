package models

import (
	"log/slog"
	"time"
)

type Client struct {
	ID        int
	Name      string
	Version   int
	Image     string
	CPU       string
	Memory    string
	Priority  float64
	SpawnedAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *Client) LogValue() slog.Value {
	if c == nil {
		return slog.StringValue("<NONE>")
	}
	return slog.IntValue(c.ID)
}
