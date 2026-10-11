package grpc

import (
	"checkgo/OneService/internal/domain"
	"checkgo/proto"
	"context"
	"encoding/json"
	"strconv"
)

type UseCase interface {
	CreateUser(ctx context.Context, user domain.User) (int64, error)
	GetUser(ctx context.Context, id string) (domain.User, error)
}
type Server struct {
	proto.UnimplementedUserServiceServer
	UseCase UseCase
}

func (s *Server) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	user, err := domain.NewUser(req.Name, req.Age, req.DateBorn)
	if err != nil {
		return nil, err
	}
	id, err := s.UseCase.CreateUser(ctx, *user)
	if err != nil {
		return nil, err
	}
	return &proto.CreateUserResponse{Id: id}, nil
}
func (s *Server) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	userId := strconv.Itoa(int(req.GetId()))
	user, err := s.UseCase.GetUser(ctx, userId)
	if err != nil {
		return nil, err
	}
	localDate, _ := json.Marshal(user.LocalDate)
	return &proto.GetUserResponse{
		Id:        user.Id,
		Name:      user.Name,
		DateBorn:  user.Birthday,
		LocalDate: string(localDate),
	}, nil
}
