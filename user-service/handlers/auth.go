package handlers

import (
	"context"
	"errors"
	"github.com/Noah-Wilderom/video-streaming/shared/crypt"
	"github.com/Noah-Wilderom/video-streaming/shared/uuid"
	"github.com/Noah-Wilderom/video-streaming/user-service/models"
	"github.com/Noah-Wilderom/video-streaming/user-service/proto/auth"
	"github.com/Noah-Wilderom/video-streaming/user-service/token"
	"gofr.dev/pkg/gofr/container"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"strings"
	"time"
)

type AuthHandler struct {
	*container.Container
	auth.UnimplementedAuthServiceServer
}

func (h *AuthHandler) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	h.Logger.Debug("Login function called")

	userRow := h.SQL.QueryRowContext(ctx, "SELECT * FROM users WHERE email = ?", request.Email)
	user, err := models.ScanToUser(userRow)
	if err != nil {
		h.Logger.Error(err)
		return nil, errors.New("Invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return nil, errors.New("Invalid credentials")
	}

	th := token.NewJWTTokenHandler()

	tokenStr, err := th.New(map[string]interface{}{
		"user": user,
	})
	if err != nil {
		h.Logger.Error(err)
		return nil, err
	}

	return &auth.LoginResponse{
		Token: tokenStr,
		User: &auth.User{
			Id:        user.Id,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, request *auth.RegisterRequest) (*auth.LoginResponse, error) {
	h.Logger.Debug("Register function called")

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	userId, err := uuid.NewV7()
	if err != nil {
		h.Logger.Error(err)
		return nil, err
	}

	user := &models.User{
		Id:        string(userId[:]),
		Name:      request.Name,
		Email:     request.Email,
		Password:  string(hash),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = h.SQL.ExecContext(ctx, "INSERT INTO users (id, name, email, password) VALUES (?, ?, ?, ?)", user.Id, user.Name, user.Email, user.Password)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			h.Logger.Error("A user with this email already exists")
			return nil, errors.New("a user with this email already exists")
		}
		h.Logger.Error(err)
		return nil, err
	}

	th := token.NewJWTTokenHandler()
	encryptedUser, err := crypt.EncryptStruct(user)
	if err != nil {
		h.Logger.Error(err)
		return nil, err
	}

	tokenStr, err := th.New(map[string][]byte{
		"user_id": encryptedUser,
	})
	if err != nil {
		h.Logger.Error(err)
		return nil, err
	}

	return &auth.LoginResponse{
		Token: tokenStr,
		User: &auth.User{
			Id:        user.Id,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (h *AuthHandler) Check(ctx context.Context, request *auth.CheckRequest) (*auth.LoginResponse, error) {
	h.Logger.Debug("Check function called")

	th := token.NewJWTTokenHandler()

	tokenValid, claims := th.Validate(request.Token)
	if !tokenValid {
		return nil, errors.New("invalid token")
	}

	if user, ok := claims["user"]; !ok {

	}

}
