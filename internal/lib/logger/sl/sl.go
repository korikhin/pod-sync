package sl

import (
	"log/slog"
	"os"
	"time"

	"github.com/korikhin/vortex-assignment/internal/models"
)

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) == 0 {
		// Метка UTC
		if a.Key == slog.TimeKey {
			a.Value = slog.TimeValue(a.Value.Time().UTC())
		}
		// Игнорировать пустое сообщение
		if a.Equal(slog.String(slog.MessageKey, "")) {
			return slog.Attr{}
		}
		// Игнорировать ошибку nil
		if a.Equal(Error(nil)) {
			return slog.Attr{}
		}
	}

	return a
}

func New() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: replaceAttr,
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func Client(c *models.Client) slog.Attr {
	return slog.Any("client", c)
}

func Component(c string) slog.Attr {
	return slog.String("component", c)
}

func Duration(d time.Duration) slog.Attr {
	return slog.Duration("duration_nanos", d)
}

func Error(err error) slog.Attr {
	return slog.Any("error", err)
}

func Operation(op string) slog.Attr {
	return slog.String("operation", op)
}

func PodOperation(po *models.PodOperation) slog.Attr {
	return slog.Any("pod_operation", po)
}

func RequestID(id string) slog.Attr {
	return slog.String("request_id", id)
}

func Signal(s os.Signal) slog.Attr {
	return slog.String("signal", s.String())
}

func Status(s models.Status) slog.Attr {
	return slog.Any("status", s)
}
