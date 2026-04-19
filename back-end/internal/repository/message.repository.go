package repository

import (
	"chat-app/back-end/internal/model"
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const maxMessageHistory = 500

type MessageRepository struct {
	rdb *redis.Client
}

func NewMessageRepository(rdb *redis.Client) *MessageRepository {
	return &MessageRepository{rdb: rdb}
}

func (r *MessageRepository) SaveMessage(ctx context.Context, msg *model.ChatMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	key := "room:messages:" + msg.RoomID
	pipe := r.rdb.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(msg.Timestamp), Member: string(data)})
	// keep the sorted set bounded
	pipe.ZRemRangeByRank(ctx, key, 0, -(maxMessageHistory + 1))
	_, err = pipe.Exec(ctx)
	return err
}

func (r *MessageRepository) GetRecentMessages(ctx context.Context, roomID string, count int64) ([]*model.ChatMessage, error) {
	results, err := r.rdb.ZRange(ctx, "room:messages:"+roomID, -count, -1).Result()
	if err != nil {
		return nil, err
	}

	msgs := make([]*model.ChatMessage, 0, len(results))
	for _, raw := range results {
		var msg model.ChatMessage
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			continue
		}
		msgs = append(msgs, &msg)
	}
	return msgs, nil
}
