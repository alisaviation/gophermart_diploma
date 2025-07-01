package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alisaviation/internal/database"
	"github.com/alisaviation/internal/gophermart/models"
)

func Test_authService_Register(t *testing.T) {
	mockJWT := &MockJWTService{
		GenerateTokenFunc: func(userID int, login string) (string, error) {
			return "generated.jwt.token", nil
		},
	}

	tests := []struct {
		name        string
		userRepo    database.UserRepository
		jwtService  *MockJWTService
		login       string
		password    string
		want        string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful registration",
			userRepo: &MockUserRepository{
				GetUserByLoginFunc: func(login string) (*models.User, error) {
					return nil, nil
				},
				CreateUserFunc: func(user models.User) error {
					return nil
				},
			},
			jwtService: mockJWT,
			login:      "validuser",
			password:   "securepassword123",
			want:       "generated.jwt.token",
			wantErr:    false,
		},
		{
			name: "login already taken",
			userRepo: &MockUserRepository{
				GetUserByLoginFunc: func(login string) (*models.User, error) {
					return &models.User{Login: "existinguser"}, nil
				},
			},
			jwtService:  mockJWT,
			login:       "existinguser",
			password:    "anypassword",
			want:        "",
			wantErr:     true,
			expectedErr: ErrLoginTaken,
		},
		{
			name: "database error on user check",
			userRepo: &MockUserRepository{
				GetUserByLoginFunc: func(login string) (*models.User, error) {
					return nil, fmt.Errorf("database connection failed")
				},
			},
			jwtService: mockJWT,
			login:      "anyuser",
			password:   "anypassword",
			want:       "",
			wantErr:    true,
		},
		{
			name: "password hashing failed",
			userRepo: &MockUserRepository{
				GetUserByLoginFunc: func(login string) (*models.User, error) {
					return nil, nil
				},
				CreateUserFunc: func(user models.User) error {
					return nil
				},
			},
			jwtService: mockJWT,
			login:      "validuser",
			password:   string([]byte{0xff}),
			want:       "",
			wantErr:    true,
		},
		{
			name: "empty password hashing failed",
			userRepo: &MockUserRepository{
				GetUserByLoginFunc: func(login string) (*models.User, error) {
					return nil, nil
				},
				CreateUserFunc: func(user models.User) error {
					return nil
				},
			},
			jwtService: mockJWT,
			login:      "validuser",
			password:   "",
			want:       "",
			wantErr:    true,
		},
		{
			name: "user creation failed",
			userRepo: &MockUserRepository{
				GetUserByLoginFunc: func(login string) (*models.User, error) {
					return nil, nil
				},
				CreateUserFunc: func(user models.User) error {
					return fmt.Errorf("creation failed")
				},
			},
			jwtService: mockJWT,
			login:      "validuser",
			password:   "goodpassword",
			want:       "",
			wantErr:    true,
		},
		{
			name: "token generation failed",
			userRepo: &MockUserRepository{
				GetUserByLoginFunc: func(login string) (*models.User, error) {
					return nil, nil
				},
				CreateUserFunc: func(user models.User) error {
					return nil
				},
			},
			jwtService: &MockJWTService{
				GenerateTokenFunc: func(userID int, login string) (string, error) {
					return "", fmt.Errorf("token error")
				},
			},
			login:    "validuser",
			password: "goodpassword",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &authService{
				userRepo:   tt.userRepo,
				jwtService: tt.jwtService,
			}
			got, err := s.Register(tt.login, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Register() expected error = %v, got %v", tt.expectedErr, err)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Register() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockUserRepository struct {
	GetUserByLoginFunc func(login string) (*models.User, error)
	CreateUserFunc     func(user models.User) error
}

func (m *MockUserRepository) GetUserByLogin(login string) (*models.User, error) {
	return m.GetUserByLoginFunc(login)
}

func (m *MockUserRepository) CreateUser(user models.User) error {
	return m.CreateUserFunc(user)
}

type MockJWTService struct {
	GenerateTokenFunc func(userID int, login string) (string, error)
}

func (m *MockJWTService) GenerateToken(userID int, login string) (string, error) {
	return m.GenerateTokenFunc(userID, login)
}
