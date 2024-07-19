package models

import "log/slog"

type Status struct {
	ID   int
	VWAP bool
	TWAP bool
	HFT  bool
}

func (s *Status) LogValue() slog.Value {
	if s == nil {
		return slog.StringValue("<NONE>")
	}
	return slog.IntValue(s.ID)
}
