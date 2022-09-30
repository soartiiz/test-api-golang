// Lucas FOLLIOT
package todo

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
)

type SessionStoreRedis struct {
	rdb *redis.Client
}

func NewSessionStoreRedis(db *redis.Client) *SessionStoreRedis {
	return &SessionStoreRedis{db}
}

func (st *SessionStoreRedis) Add(userID uuid.UUID, token string) error {
	err := st.rdb.Set(context.Background(), token, userID.String(), 0).Err()
	if err != nil {
		return err
	}

	val, err := st.rdb.Get(context.Background(), userID.String()).Result()
	if err != nil {
		return err
	}

	fmt.Println(val)

	return nil
}

func (st *SessionStoreRedis) Revoke(token string) error {
	val, err := st.rdb.Get(context.Background(), token).Result()
	if err != nil {
		return err
	}

	v, err := st.rdb.Del(context.Background(), val).Result()
	if err != nil {
		return err
	}

	fmt.Println(v)

	return nil
}

func (st *SessionStoreRedis) FindByToken(token string) (userID uuid.UUID, err error) {
	userId, err := st.rdb.Get(context.Background(), token).Result()

	return uuid.FromStringOrNil(userId), err
}
