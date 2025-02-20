package dto

import (
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lightlink/auth-service/internal/session/domain/entity"
	"github.com/lightlink/auth-service/internal/session/domain/model"
)

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func createJWT(signupRequest *SignupRequest, ttl time.Time, userID uint) (string, error) {
	jwtTokenKey := []byte(os.Getenv("TOKEN_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]string{
			"username": signupRequest.Username,
			"id":       strconv.Itoa(int(userID)),
		},
		"iat": time.Now().Unix(),
		"exp": ttl,
	})
	tokenString, err := token.SignedString(jwtTokenKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func createSignedSession(signupRequest *SignupRequest, userID uint, accessTokenTTL time.Time, refreshTokenTTL time.Time) (*entity.Session, error) {
	accessToken, err := createJWT(signupRequest, accessTokenTTL, userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := createJWT(signupRequest, refreshTokenTTL, userID)
	if err != nil {
		return nil, err
	}

	return &entity.Session{
		JWTAccess:        accessToken,
		JWTRefresh:       refreshToken,
		UserID:           userID,
		Username:         signupRequest.Username,
		AccessExpiresAt:  accessTokenTTL,
		RefreshExpiresAt: refreshTokenTTL,
	}, nil
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

func SignupRequestToEntity(signupRequest *SignupRequest, userID uint, accessTokenTTL time.Time, refreshTokenTTL time.Time) (*entity.Session, error) {
	signedSessionEntity, err := createSignedSession(signupRequest, userID, accessTokenTTL, refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return signedSessionEntity, nil
}
