package grpc

import (
	"context"

	"github.com/lightlink/auth-service/internal/user/domain/dto"
	"github.com/lightlink/auth-service/internal/user/domain/entity"
	proto "github.com/lightlink/auth-service/protogen/user"
)

type UserGrpcRepository struct {
	Client proto.UserServiceClient
}

func NewUserGrpcRepository(client *proto.UserServiceClient) *UserGrpcRepository {
	return &UserGrpcRepository{
		Client: *client,
	}
}

func (repo *UserGrpcRepository) Create(userEntity *entity.User) (*dto.UserTransfer, error) {
	createUserRequest := dto.UserEntityToCreateRequest(userEntity)
	userResponseProto, err := repo.Client.CreateUser(context.Background(), createUserRequest)
	if err != nil {
		return nil, err
	}

	createdUser := dto.GetUserResponseToTransfer(userResponseProto)

	return createdUser, nil
}

func (repo *UserGrpcRepository) GetById(id uint) (*dto.UserTransfer, error) {
	getUserByIdRequest := &proto.GetUserByIdRequest{
		Id: uint32(id),
	}

	userResponseProto, err := repo.Client.GetUserById(context.Background(), getUserByIdRequest)
	if err != nil {
		return nil, err
	}

	userModel := dto.GetUserResponseToTransfer(userResponseProto)

	return userModel, nil
}

func (repo *UserGrpcRepository) GetByUsername(username string) (*dto.UserTransfer, error) {
	getUserByUsernameRequest := &proto.GetUserByUsernameRequest{
		Username: username,
	}

	userResponseProto, err := repo.Client.GetUserByUsername(context.Background(), getUserByUsernameRequest)
	if err != nil {
		return nil, err
	}

	userModel := dto.GetUserResponseToTransfer(userResponseProto)

	return userModel, nil
}
