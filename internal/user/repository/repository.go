package repository

import (
	"github.com/lightlink/auth-service/internal/user/domain/dto"
	"github.com/lightlink/auth-service/internal/user/domain/entity"
)

type UserRepositoryI interface {
	Create(userEntity *entity.User) (*dto.UserTransfer, error)
	GetById(id uint) (*dto.UserTransfer, error)
	GetByUsername(username string) (*dto.UserTransfer, error)
}
