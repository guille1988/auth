package services

import (
	"auth/internal/infrastructure/redis"
	"context"
	"time"
)

const (
	sessionKeyPrefix      = "auth:token:"
	userSessionsKeyPrefix = "auth:user_sessions:"
)

// SessionKey builds the Redis key holding one refresh-token session.
func SessionKey(refreshTokenUUID string) string {
	return sessionKeyPrefix + refreshTokenUUID
}

func userSessionsKey(userUUID string) string {
	return userSessionsKeyPrefix + userUUID
}

/*
SessionIndex maintains the per-user secondary index of live sessions: a
sorted set keyed by user UUID whose members are refresh-token UUIDs scored
by their expiry timestamp. It exists because the access token carries only
the user UUID, so "is this token's user still logged in somewhere?" needs a
reverse lookup that the primary auth:token:* keys cannot answer without a
full scan. The index is best-effort: the session key remains the source of
truth, and a missed index writing self-heals through the client's refresh.
*/
type SessionIndex struct {
	redisRepository *redis.Repository
}

func NewSessionIndex(redisRepository *redis.Repository) *SessionIndex {
	return &SessionIndex{redisRepository: redisRepository}
}

/*
Add registers a session under its user, refreshing the index TTL so an
abandoned index dies together with its longest-lived session.
*/
func (index *SessionIndex) Add(ctx context.Context, userUUID, refreshTokenUUID string, expiresAt time.Time) error {
	return index.redisRepository.AddToSortedSet(
		ctx,
		userSessionsKey(userUUID),
		refreshTokenUUID,
		float64(expiresAt.Unix()),
		time.Until(expiresAt),
	)
}

// Remove drops one session from the user's index.
func (index *SessionIndex) Remove(ctx context.Context, userUUID, refreshTokenUUID string) error {
	return index.redisRepository.RemoveFromSortedSet(ctx, userSessionsKey(userUUID), refreshTokenUUID)
}

// HasLiveSession reports whether the user has at least one unexpired session.
func (index *SessionIndex) HasLiveSession(ctx context.Context, userUUID string) (bool, error) {
	count, err := index.redisRepository.CountLiveMembers(ctx, userSessionsKey(userUUID), time.Now())

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
