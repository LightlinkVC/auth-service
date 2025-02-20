package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/lightlink/auth-service/internal/session/domain/dto"
	"github.com/lightlink/auth-service/internal/session/domain/entity"
	"github.com/lightlink/auth-service/internal/session/domain/model"
)

type SessionRedisRepository struct {
	redisConn redis.Conn
	mu        *sync.Mutex
}

func NewSessionRedisRepository(conn redis.Conn) *SessionRedisRepository {
	return &SessionRedisRepository{
		redisConn: conn,
		mu:        &sync.Mutex{},
	}
}

func (repo *SessionRedisRepository) Create(sessionEntity *entity.Session) (*model.Session, error) {
	mkey := "sessions:" + strconv.Itoa(int(sessionEntity.UserID))

	sessionModel := dto.SessionEntityToModel(sessionEntity)
	sessionSerialized, err := json.Marshal(sessionModel)
	if err != nil {
		return nil, err
	}

	ttl := int(time.Until(sessionModel.RefreshExpiresAt))

	repo.mu.Lock()
	result, err := redis.String(repo.redisConn.Do("SET", mkey, sessionSerialized, "EX", ttl))
	repo.mu.Unlock()

	if err != nil {
		fmt.Println("here 2")
		return nil, err
	}

	if result != "OK" {
		return nil, fmt.Errorf("unexpected Redis response: %v", result)
	}

	return sessionModel, nil
}

func (repo *SessionRedisRepository) Check(userID uint) (*model.Session, error) {
	mkey := "sessions:" + strconv.Itoa(int(userID))

	repo.mu.Lock()
	bytes, err := redis.Bytes(repo.redisConn.Do("GET", mkey))
	repo.mu.Unlock()

	if err == redis.ErrNil {
		return nil, entity.ErrNoSession
	}

	if err != nil {
		return nil, err
	}

	session := &model.Session{}
	err = json.Unmarshal(bytes, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (sm *SessionRedisRepository) Delete(userID uint) error {
	mkey := "sessions:" + strconv.Itoa(int(userID))

	sm.mu.Lock()
	_, err := redis.Int(sm.redisConn.Do("DEL", mkey))
	sm.mu.Unlock()

	if err == redis.ErrNil {
		return entity.ErrNoSession
	}

	if err != nil {
		return err
	}

	return nil
}
