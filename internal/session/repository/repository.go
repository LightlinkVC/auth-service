package repository

import (
	"github.com/lightlink/auth-service/internal/session/domain/entity"
	"github.com/lightlink/auth-service/internal/session/domain/model"
)

type SessionRepositoryI interface {
	Create(sessionEntity *entity.Session) (*model.Session, error)
	Check(userID uint) (*model.Session, error)
	Delete(userID uint) error
}
