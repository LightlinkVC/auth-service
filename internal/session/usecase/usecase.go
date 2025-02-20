package usecase

import (
	"time"

	sessionDTO "github.com/lightlink/auth-service/internal/session/domain/dto"
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

	session, err := sessionDTO.SignupRequestToEntity(
		signupRequest,
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

	createdSessionModel, err := uc.sessionRepo.Create(session)
	if err != nil {
		return nil, err
	}

	createdSessionEntity := sessionDTO.SessionModelToEntity(createdSessionModel)

	return createdSessionEntity, nil
}

/*TODO*/
// func (uc *SessionUsecase) Create(signupRequest *sessionDTO.SignupRequest) (*sessionEntity.Session, error) {
// 	sessionEntity, err := sessionDTO.SignupRequestToEntity(
// 		signupRequest,

// 		time.Now().Add(15*time.Minute), /*TODO*/
// 		time.Now().Add(24*time.Hour),   /*TODO*/
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	_, err = uc.sessionRepo.Check(uint(sessionEntity.UserID))
// 	if err != nil {
// 		return nil, err
// 	}

// 	createdSessionModel, err := uc.sessionRepo.Create(sessionEntity)
// 	if err != nil {
// 		return nil, err
// 	}

// 	createdSessionEntity := sessionDTO.SessionModelToEntity(createdSessionModel)

// 	return createdSessionEntity, nil
// }

// func (uc *SessionUsecase) Check(userID uint) (*sessionEntity.Session, error) {
// 	sessionModel, err := uc.sessionRepo.Check(userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	sessionEntity := sessionDTO.SessionModelToEntity(sessionModel)

// 	return sessionEntity, nil
// }
