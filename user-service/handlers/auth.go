package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/shared/crypt"
	"github.com/Noah-Wilderom/video-streaming/shared/token"
	"github.com/Noah-Wilderom/video-streaming/user-service/models"
	"github.com/Noah-Wilderom/video-streaming/user-service/proto/auth"
	"github.com/google/uuid"
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
	h.Logger.Info("Login function called")

	userRow := h.SQL.QueryRowContext(ctx, "SELECT * FROM users WHERE email = ?", request.Email)
	user, err := models.ScanToUser(userRow)
	if err != nil {
		h.Logger.Error(err.Error())
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		fmt.Println("wrong password")
		return nil, errors.New("invalid credentials")
	}

	th := token.NewJWTTokenHandler()
	encryptedUser, err := crypt.EncryptStructBase64(user)
	if err != nil {
		return nil, err
	}

	tokenStr, err := th.New(map[string]string{
		"user": encryptedUser,
	})
	if err != nil {
		h.Logger.Error(err.Error())
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

	fmt.Println(request.GetPassword())
	hash, err := bcrypt.GenerateFromPassword([]byte(request.GetPassword()), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	userId, err := uuid.NewV7()
	if err != nil {
		h.Logger.Error(err)
		return nil, err
	}

	user := &models.User{
		Id:        userId.String(),
		Name:      request.GetName(),
		Email:     request.GetEmail(),
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
	encryptedUser, err := crypt.EncryptStructBase64(user)
	if err != nil {
		h.Logger.Error(err)
		return nil, err
	}

	tokenStr, err := th.New(map[string]string{
		"user": encryptedUser,
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
	h.Logger.Debug("check function called")

	th := token.NewJWTTokenHandler()

	tokenValid, data := th.Validate(request.Token)
	if !tokenValid {
		return nil, errors.New("invalid token")
	}

	encryptedUser, ok := data["user"].(string)
	if !ok {
		h.Logger.Error("invalid user data")
		return nil, errors.New("invalid token")
	}

	var user models.User
	err := crypt.DecryptStructBase64(encryptedUser, &user)
	if err != nil {
		h.Logger.Error(errors.Join(errors.New("invalid user data encryption error"), err))
		return nil, err
	}

	return &auth.LoginResponse{
		Token: request.Token,
		User: &auth.User{
			Id:        user.Id,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.CreatedAt),
		},
	}, nil
}
