package usecase

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	sessionDTO "github.com/lightlink/auth-service/internal/session/domain/dto"
	"github.com/lightlink/auth-service/internal/session/domain/entity"
	sessionEntity "github.com/lightlink/auth-service/internal/session/domain/entity"
	sessionRepo "github.com/lightlink/auth-service/internal/session/repository"
	userDTO "github.com/lightlink/auth-service/internal/user/domain/dto"
	userEntity "github.com/lightlink/auth-service/internal/user/domain/entity"
	userRepo "github.com/lightlink/auth-service/internal/user/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SessionUsecaseI interface {
	Signup(signupRequest *sessionDTO.SignupRequest) (*sessionEntity.Session, error)
	Login(loginRequest *sessionDTO.LoginRequest) (*sessionEntity.Session, error)
	RefreshSession(refreshToken *jwt.Token) (*sessionEntity.Session, error)
	/*TODO*/
	// Create(signupRequest *sessionDTO.SignupRequest) (*sessionEntity.Session, error)
	// Check(userID uint) (*sessionEntity.Session, error)
}

type SessionUsecase struct {
	sessionRepo sessionRepo.SessionRepositoryI
	userRepo    userRepo.UserRepositoryI
}

func NewSessionUsecase(sessionRepository sessionRepo.SessionRepositoryI, userRepository userRepo.UserRepositoryI) *SessionUsecase {
	return &SessionUsecase{
		sessionRepo: sessionRepository,
		userRepo:    userRepository,
	}
}

func (uc *SessionUsecase) Signup(signupRequest *sessionDTO.SignupRequest) (*sessionEntity.Session, error) {
	_, err := uc.userRepo.GetByUsername(signupRequest.Username)
	if err == nil {
		return nil, userEntity.ErrAlreadyCreated
	}

	if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
		return nil, err
	}

	userEntity, err := userDTO.SignupRequestToEntity(signupRequest)
	if err != nil {
		return nil, err
	}

	createdUser, err := uc.userRepo.Create(userEntity)
	if err != nil {
		return nil, err
	}

	authDTO := sessionDTO.SignupRequestToAuthCredentialsDTO(signupRequest)
	session, err := formSignedSession(
		authDTO.Username,
		createdUser.Id,
		time.Now().Add(15*time.Minute), /*TODO*/
		time.Now().Add(24*time.Hour),   /*TODO*/
	)
	if err != nil {
		return nil, err
	}

	sessionCheck, err := uc.sessionRepo.Check(uint(session.UserID))
	if err == sessionEntity.ErrAlreadyCreated {
		return sessionDTO.SessionModelToEntity(sessionCheck), nil
	}

	if err != sessionEntity.ErrNoSession {
		return nil, err
	}

	createdSessionModel, err := uc.sessionRepo.Set(session)
	if err != nil {
		return nil, err
	}

	createdSessionEntity := sessionDTO.SessionModelToEntity(createdSessionModel)

	return createdSessionEntity, nil
}

func (uc *SessionUsecase) Login(loginRequest *sessionDTO.LoginRequest) (*sessionEntity.Session, error) {
	user, err := uc.userRepo.GetByUsername(loginRequest.Username)

	if st, ok := status.FromError(err); !ok || st.Code() == codes.NotFound {
		return nil, errors.New("should signup first")
	}

	if err != nil {
		return nil, err
	}

	_, err = uc.sessionRepo.Check(user.Id)
	if err == sessionEntity.ErrAlreadyCreated {
		return nil, sessionEntity.ErrAlreadyCreated
	}

	authDTO := sessionDTO.LoginRequestToAuthCredentialsDTO(loginRequest)
	session, err := formSignedSession(
		authDTO.Username,
		user.Id,
		time.Now().Add(1*time.Minute), /*TODO*/
		time.Now().Add(24*time.Hour),  /*TODO*/
	)
	if err != nil {
		return nil, err
	}

	_, err = uc.sessionRepo.Check(uint(session.UserID))
	if err == nil {
		return nil, sessionEntity.ErrAlreadyCreated
	}

	if err != sessionEntity.ErrNoSession {
		return nil, err
	}

	createdSessionModel, err := uc.sessionRepo.Set(session)
	if err != nil {
		return nil, err
	}

	createdSessionEntity := sessionDTO.SessionModelToEntity(createdSessionModel)

	return createdSessionEntity, nil
}

func (uc *SessionUsecase) RefreshSession(refreshToken *jwt.Token) (*sessionEntity.Session, error) {
	claims, ok := refreshToken.Claims.(jwt.MapClaims)
	if !ok || !refreshToken.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claimsUser, ok := claims["user"].(map[string]interface{})
	if !ok {
		/*Handle*/
		fmt.Println("User claims are missing")
		return nil, errors.New("User claims are missing")
	}

	userIDString, ok := claimsUser["id"].(string)
	if !ok {
		/*Handle*/
		fmt.Println("Couldn't parse user id")
		return nil, errors.New("Couldn't parse user id")
	}

	userID64, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil || userID64 == 0 {
		/*Handle*/
		fmt.Println("Encounter user id <= 0")
		return nil, errors.New("Encounter user id <= 0")
	}

	userID := uint(userID64)
	username, ok := claimsUser["username"].(string)
	if !ok {
		/*Handle*/
		fmt.Println("Couldn't cast username to string")
		return nil, errors.New("Couldn't cast username to string")
	}

	_, err = uc.sessionRepo.Check(userID)
	if err != nil {
		return nil, err
	}

	updatedSessionEntity, err := formSignedSession(
		username,
		userID,
		time.Now().Add(15*time.Minute), /*TODO*/
		time.Now().Add(24*time.Hour),   /*TODO*/
	)
	if err != nil {
		return nil, err
	}

	_, err = uc.sessionRepo.Set(updatedSessionEntity)
	if err != nil {
		return nil, err
	}

	return updatedSessionEntity, nil
}

func createJWT(username string, ttl time.Time, userID uint) (string, error) {
	jwtTokenKey := []byte(os.Getenv("TOKEN_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]string{
			"username": username,
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

func formSignedSession(username string, userID uint, accessTokenTTL time.Time, refreshTokenTTL time.Time) (*entity.Session, error) {
	accessToken, err := createJWT(username, accessTokenTTL, userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := createJWT(username, refreshTokenTTL, userID)
	if err != nil {
		return nil, err
	}

	return &entity.Session{
		JWTAccess:        accessToken,
		JWTRefresh:       refreshToken,
		UserID:           userID,
		Username:         username,
		AccessExpiresAt:  accessTokenTTL,
		RefreshExpiresAt: refreshTokenTTL,
	}, nil
}
