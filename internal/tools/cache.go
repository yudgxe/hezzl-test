package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/yudgxe/hezzl-test/internal/database/sqlc"
	"golang.org/x/net/context"
)

// Cache - простая обретка над редис клиентом.
// Умеет сохранять/получать структуры в формате json.
// Не лучший формат в плане производительности, выбран для простоты.
type Cache struct {
	*redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{
		Client: client,
	}
}

func (c *Cache) setStruct(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, b, expiration).Err()
}

func (c *Cache) getStruct(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// ScanKey - проверяет существует ли кей, возращает первый подходящий.
func (c *Cache) ScanKey(ctx context.Context, match string) (string, bool, error) {
	var cursor uint64
	var keys []string
	var err error

	for {
		keys, cursor, err = c.Scan(context.Background(), cursor, match, 0).Result()
		if err != nil {
			return "", false, err
		}
		for _, key := range keys {
			return key, true, nil
		}
		if cursor == 0 {
			break
		}
	}
	return "", false, nil
}

func (c *Cache) SetGood(ctx context.Context, good sqlc.Good, expiration time.Duration, ifexist bool) error {
	return c.setGood(ctx, fmt.Sprintf("%d:%d:*", good.ID, good.ProjectID), good, expiration, ifexist)
}

func (c *Cache) SetGoodWihtPagination(ctx context.Context, good sqlc.Good, expiration time.Duration, ifexist bool, pagination int) error {
	return c.setGood(ctx, fmt.Sprintf("%d:%d:%d", good.ID, good.ProjectID, pagination), good, expiration, ifexist)
}

func (c *Cache) setGood(ctx context.Context, key string, good sqlc.Good, expiration time.Duration, ifexist bool) error {
	if ifexist {
		gk, ok, err := c.ScanKey(ctx, key)
		if err != nil {
			return err
		}
		if ok {
			if err := c.setStruct(ctx, gk, good, expiration); err != nil {
				return err
			}
			log.Info().Interface("good", good).Msg("updated cash")
		}
	} else {
		if err := c.setStruct(ctx, key, good, expiration); err != nil {
			return err
		}
		log.Info().Interface("good", good).Msg("set cash")
	}

	return nil
}

type GetGoodsWithPaginationReponse struct {
	pagination  Pagination
	mergeIndex  int
	hasNotFound bool
}

func (ggwpr *GetGoodsWithPaginationReponse) HasNotFound() bool { return ggwpr.hasNotFound }

func (ggwpr *GetGoodsWithPaginationReponse) MergeIndex() int { return ggwpr.mergeIndex }

func (ggwpr *GetGoodsWithPaginationReponse) Pagination() Pagination { return ggwpr.pagination }

func (c *Cache) GetGoodsWithPagination(ctx context.Context, pagination Pagination) ([]sqlc.Good, *GetGoodsWithPaginationReponse, error) {
	var good sqlc.Good
	nf := make([]int, 0)
	response := make([]sqlc.Good, 0)
	for i := pagination.Offset + 1; i < pagination.Offset+pagination.Limit+1; i++ {
		key, ok, err := c.ScanKey(context.Background(), fmt.Sprintf("*:*:%d", i))
		if err != nil {
			log.Printf("%s\n", err)
		}
		if !ok {
			nf = append(nf, i)
		} else {
			if err := c.getStruct(ctx, key, &good); err != nil {
				// TODO: add wraper for erros
				log.Printf("%s\n", err)

				nf = append(nf, i)
				continue
			}
			log.Info().Interface("good", good).Msg("get cash")
			response = append(response, good)
		}

	}
	ggwpr := new(GetGoodsWithPaginationReponse)
	// Ставим новый offset и limit.
	// Пример: offset = 5, limit = 5 -> необходимо найти товары с offset'ом [6, 7, 8, 9, 10].
	// Допустим в кеше уже есть 6 и 10, получаем новый оффсет = 7 - 1, а лимит = 9 - 6 -> offset = 6, а limit = 3 -> получаем с базы [7, 8, 9].
	// Так же считает индекс для мерджа товаров с базы в найденых с кеша, index = new_offset - old_offset.
	// Пример: new_offset - old_offset = 6 - 5 = 1, когда в кеше [6, 10], а ответе с базы [7, 8, 9]
	if len(nf) != 0 {
		of := pagination.Offset
		ggwpr.pagination.Offset = nf[0] - 1
		ggwpr.pagination.Limit = nf[len(nf)-1] - ggwpr.pagination.Offset
		ggwpr.mergeIndex = pagination.Offset - of
		ggwpr.hasNotFound = true
	}

	return response, ggwpr, nil
}
