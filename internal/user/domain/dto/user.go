package dto

import (
	"github.com/lightlink/auth-service/internal/session/domain/dto"
	"github.com/lightlink/auth-service/internal/user/domain/entity"
	proto "github.com/lightlink/auth-service/protogen/user"
	"golang.org/x/crypto/bcrypt"
)

type UserTransfer struct {
	Id       uint
	Username string
}

func GetUserResponseToTransfer(getResponse *proto.GetUserResponse) *UserTransfer {
	return &UserTransfer{
		Id:       uint(getResponse.Id),
		Username: getResponse.Username,
	}
}

func UserEntityToCreateRequest(userEntity *entity.User) *proto.CreateUserRequest {
	return &proto.CreateUserRequest{
		Username:     userEntity.Username,
		PasswordHash: userEntity.PasswordHash,
	}
}

func SignupRequestToEntity(signupRequest *dto.SignupRequest) (*entity.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signupRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &entity.User{
		Username:     signupRequest.Username,
		PasswordHash: string(hashedPassword),
	}, nil
}
