package clickhouse

import (
	"time"

	"github.com/yudgxe/hezzl-test/internal/database/sqlc"
)

type Good struct {
	sqlc.Good
	EventTime time.Time
}

func FromGoodSQLC(good sqlc.Good) Good {
	return Good{
		Good:      good,
		EventTime: time.Now(),
	}
}
