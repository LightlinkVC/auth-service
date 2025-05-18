package dto

import (
	"github.com/lightlink/auth-service/internal/session/domain/entity"
	"github.com/lightlink/auth-service/internal/session/domain/model"
)

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SessionEntityToModel(sessionEntity *entity.Session) *model.Session {
	return &model.Session{
		JWTAccess:        sessionEntity.JWTAccess,
		JWTRefresh:       sessionEntity.JWTRefresh,
		UserID:           sessionEntity.UserID,
		Username:         sessionEntity.Username,
		AccessExpiresAt:  sessionEntity.AccessExpiresAt,
		RefreshExpiresAt: sessionEntity.RefreshExpiresAt,
	}
}

func SessionModelToEntity(sessionModel *model.Session) *entity.Session {
	return &entity.Session{
		JWTAccess:        sessionModel.JWTAccess,
		JWTRefresh:       sessionModel.JWTRefresh,
		UserID:           sessionModel.UserID,
		Username:         sessionModel.Username,
		AccessExpiresAt:  sessionModel.AccessExpiresAt,
		RefreshExpiresAt: sessionModel.RefreshExpiresAt,
	}
}
