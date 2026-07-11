package redis

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	client *redis.Client
}

func NewRepository(client *redis.Client) *Repository {
	return &Repository{client: client}
}

func (repository *Repository) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)

	if err != nil {
		return err
	}

	return repository.client.Set(ctx, key, data, expiration).Err()
}

func (repository *Repository) Get(ctx context.Context, key string, dest any) error {
	data, err := repository.client.Get(ctx, key).Bytes()

	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

func (repository *Repository) Delete(ctx context.Context, key string) error {
	return repository.client.Del(ctx, key).Err()
}

/*
GetDel atomically retrieves and deletes the key, so concurrent callers
presenting the same one-time token cannot both succeed (only the first
GETDEL wins; the rest see a miss).
*/
func (repository *Repository) GetDel(ctx context.Context, key string, dest any) error {
	data, err := repository.client.GetDel(ctx, key).Bytes()

	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

/*
AddToSortedSet inserts the member with the given score and refreshes the
set's TTL, pipelined into one round trip. ZADD and EXPIRE always travel
together, so a caller can never leave an immortal set behind.
*/
func (repository *Repository) AddToSortedSet(ctx context.Context, key, member string, score float64, ttl time.Duration) error {
	pipeline := repository.client.TxPipeline()
	pipeline.ZAdd(ctx, key, redis.Z{Score: score, Member: member})
	pipeline.Expire(ctx, key, ttl)

	_, err := pipeline.Exec(ctx)

	return err
}

/*
RemoveFromSortedSet removes the member from the sorted set; missing
members and missing keys are no-ops.
*/
func (repository *Repository) RemoveFromSortedSet(ctx context.Context, key, member string) error {
	return repository.client.ZRem(ctx, key, member).Err()
}

/*
CountLiveMembers counts the members whose score is strictly after now,
purging the expired ones first (pipelined). The purge lives inside the
count on purpose: a caller that skipped it would count expired members
as live.
*/
func (repository *Repository) CountLiveMembers(ctx context.Context, key string, now time.Time) (int64, error) {
	pipeline := repository.client.TxPipeline()
	pipeline.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(now.Unix(), 10))
	count := pipeline.ZCard(ctx, key)

	if _, err := pipeline.Exec(ctx); err != nil {
		return 0, err
	}

	return count.Val(), nil
}
