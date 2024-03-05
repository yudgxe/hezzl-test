package tools

import (
	"fmt"
	"strings"

	"github.com/yudgxe/hezzl-test/internal/model/clickhouse"
	"golang.org/x/net/context"
)

type Store interface {
	Exec(ctx context.Context, query string, args ...any) error
}

type GoodSender struct {
	store Store
}

func NewGoodSender(store Store) *GoodSender {
	return &GoodSender{
		store: store,
	}
}

func (s *GoodSender) Send(data interface{}) error {
	switch v := data.(type) {
	case []clickhouse.Good:
		if len(v) <= 0 {
			return nil
		}

		var sb strings.Builder

		// TODO: reflect
		// количество полей у структуры.
		countField := 8

		args := make([]any, 0, len(v)*countField)

		sb.WriteString("INSERT INTO goods VALUES ")
		for _, good := range v {
			size := len(args) + 1
			subquery := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d), ", size, size+1, size+2, size+3, size+4, size+5, size+6, size+7)
			args = append(args, good.ID, good.ProjectID, good.Name, good.Description, good.Priority, good.Removed, good.CreatedAt, good.EventTime)

			sb.WriteString(subquery)
		}
		query := sb.String()
		query = query[:len(query)-2]

		if err := s.store.Exec(context.Background(), query, args...); err != nil {
			return err
		}
	}

	return nil
}
