package tests

import (
	"errors"
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/internal/gophermart/models"
	"github.com/alisaviation/internal/gophermart/services"
	"github.com/alisaviation/internal/tests/mocks"
)

func Test_authService_Register(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mocks.MockUserRepository, *mocks.MockJWTService)
		login       string
		password    string
		want        string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful registration",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
				mur.On("GetUserByLogin", "validuser").Return((*models.User)(nil), nil)
				mur.On("CreateUser", mock.AnythingOfType("models.User")).Return(1, nil)
				mjwt.On("GenerateToken", 1, "validuser").Return("generated.jwt.token", nil)
			},
			login:    "validuser",
			password: "securepassword123",
			want:     "generated.jwt.token",
			wantErr:  false,
		},
		{
			name: "login already taken",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
				mur.On("GetUserByLogin", "existinguser").Return(&models.User{Login: "existinguser"}, nil)
			},
			login:       "existinguser",
			password:    "anypassword",
			want:        "",
			wantErr:     true,
			expectedErr: services.ErrLoginTaken,
		},
		{
			name: "database error on user check",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
				mur.On("GetUserByLogin", "anyuser").Return((*models.User)(nil), fmt.Errorf("database connection failed"))
			},
			login:    "anyuser",
			password: "anypassword",
			want:     "",
			wantErr:  true,
		},
		{
			name: "password hashing failed",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
			},
			login:    "validuser",
			password: string([]byte{0xff}),
			want:     "",
			wantErr:  true,
		},
		{
			name: "empty password",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
			},
			login:    "validuser",
			password: "",
			want:     "",
			wantErr:  true,
		},
		{
			name: "user creation failed",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
				mur.On("GetUserByLogin", "validuser").Return((*models.User)(nil), nil)
				mur.On("CreateUser", mock.AnythingOfType("models.User")).Return(0, fmt.Errorf("creation failed"))
			},
			login:    "validuser",
			password: "goodpassword",
			want:     "",
			wantErr:  true,
		},
		{
			name: "token generation failed",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
				mur.On("GetUserByLogin", "validuser").Return((*models.User)(nil), nil)
				mur.On("CreateUser", mock.AnythingOfType("models.User")).Return(1, nil)
				mjwt.On("GenerateToken", 1, "validuser").Return("", fmt.Errorf("token error"))
			},
			login:    "validuser",
			password: "goodpassword",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mocks.MockUserRepository{}
			mockJWT := &mocks.MockJWTService{}

			if tt.setupMock != nil {
				tt.setupMock(mockUserRepo, mockJWT)
			}

			s := &services.AuthStructService{
				UserRepo:   mockUserRepo,
				JwtService: mockJWT,
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

			mockUserRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func Test_authService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockUserRepository, *mocks.MockJWTService)
		login       string
		password    string
		want        string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful login",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
				mur.On("GetUserByLogin", "validuser").Return(&models.User{
					ID:           1,
					Login:        "validuser",
					PasswordHash: string(hashedPassword),
				}, nil)
				mjwt.On("GenerateToken", 1, "validuser").Return("generated.jwt.token", nil)
			},
			login:    "validuser",
			password: "correctpassword",
			want:     "generated.jwt.token",
			wantErr:  false,
		},
		{
			name: "invalid credentials - wrong password",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
				mur.On("GetUserByLogin", "validuser").Return(&models.User{
					Login:        "validuser",
					PasswordHash: string(hashedPassword),
				}, nil)
			},
			login:       "validuser",
			password:    "wrongpassword",
			want:        "",
			wantErr:     true,
			expectedErr: services.ErrInvalidCredentials,
		},
		{
			name: "invalid credentials - user not found",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
				mur.On("GetUserByLogin", "nonexistent").Return((*models.User)(nil), nil)
			},
			login:       "nonexistent",
			password:    "anypassword",
			want:        "",
			wantErr:     true,
			expectedErr: services.ErrInvalidCredentials,
		},
		{
			name: "empty login",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
			},
			login:       "",
			password:    "anypassword",
			want:        "",
			wantErr:     true,
			expectedErr: services.ErrInvalidCredentials,
		},
		{
			name: "empty password",
			setupMock: func(mur *mocks.MockUserRepository, mjwt *mocks.MockJWTService) {
			},
			login:       "validuser",
			password:    "",
			want:        "",
			wantErr:     true,
			expectedErr: services.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mocks.MockUserRepository{}
			mockJWT := &mocks.MockJWTService{}

			if tt.setupMock != nil {
				tt.setupMock(mockUserRepo, mockJWT)
			}

			s := &services.AuthStructService{
				UserRepo:   mockUserRepo,
				JwtService: mockJWT,
			}
			got, err := s.Login(tt.login, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Login() expected error = %v, got %v", tt.expectedErr, err)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Login() = %v, want %v", got, tt.want)
			}

			mockUserRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}
