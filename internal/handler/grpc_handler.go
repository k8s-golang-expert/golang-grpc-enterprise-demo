package handler

import (
	"context"

	"golang-grpc-enterprise-demo/internal/service"
	userv1 "golang-grpc-enterprise-demo/proto/gen/user/v1"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCHandler struct {
	userv1.UnimplementedUserServiceServer
	svc service.UserService
}

func NewGRPCHandler(svc service.UserService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	ctx, span := otel.Tracer("grpc").Start(ctx, "GetUser")
	defer span.End()

	user, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &userv1.GetUserResponse{
		User: &userv1.User{
			Id:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
		},
	}, nil
}

func (h *GRPCHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	ctx, span := otel.Tracer("grpc").Start(ctx, "CreateUser")
	defer span.End()

	user, err := h.svc.Create(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &userv1.CreateUserResponse{
		User: &userv1.User{
			Id:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
		},
	}, nil
}

func (h *GRPCHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	ctx, span := otel.Tracer("grpc").Start(ctx, "ListUsers")
	defer span.End()

	users, total, err := h.svc.List(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbUsers := make([]*userv1.User, len(users))
	for i, u := range users {
		pbUsers[i] = &userv1.User{
			Id:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: timestamppb.New(u.CreatedAt),
		}
	}
	return &userv1.ListUsersResponse{
		Users: pbUsers,
		Total: int32(total),
	}, nil
}
