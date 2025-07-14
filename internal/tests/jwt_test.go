package tests

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/alisaviation/internal/gophermart/services"
)

func TestJWTService_GenerateToken(t *testing.T) {
	validKey := []byte("test_secret_key")
	validIssuer := "test_issuer"

	type fields struct {
		secretKey []byte
		issuer    string
	}
	type args struct {
		userID int
		login  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantLen int // ожидаемая минимальная длина токена
		wantErr bool
	}{
		{
			name: "successful token generation",
			fields: fields{
				secretKey: validKey,
				issuer:    validIssuer,
			},
			args: args{
				userID: 1,
				login:  "testuser",
			},
			wantLen: 50, // JWT токены обычно длинные
			wantErr: false,
		},
		{
			name: "empty issuer",
			fields: fields{
				secretKey: validKey,
				issuer:    "",
			},
			args: args{
				userID: 1,
				login:  "testuser",
			},
			wantLen: 50,
			wantErr: false,
		},
		{
			name: "empty login",
			fields: fields{
				secretKey: validKey,
				issuer:    validIssuer,
			},
			args: args{
				userID: 1,
				login:  "",
			},
			wantLen: 50,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &services.JWTService{
				SecretKey: tt.fields.secretKey,
				Issuer:    tt.fields.issuer,
			}
			got, err := s.GenerateToken(tt.args.userID, tt.args.login)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) < tt.wantLen {
				t.Errorf("GenerateToken() got too short token, len = %d, want at least %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	validKey := []byte("test_secret_key")
	validIssuer := "test_issuer"
	invalidKey := []byte("invalid_secret_key")
	invalidIssuer := "invalid_issuer"

	// Генерируем валидный токен для тестов
	validService := &services.JWTService{
		SecretKey: validKey,
		Issuer:    validIssuer,
	}
	validToken, _ := validService.GenerateToken(1, "testuser")

	type fields struct {
		secretKey []byte
		issuer    string
	}
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *services.Claims
		wantErr bool
	}{
		{
			name: "valid token",
			fields: fields{
				secretKey: validKey,
				issuer:    validIssuer,
			},
			args: args{
				tokenString: validToken,
			},
			want: &services.Claims{
				UserID: 1,
				Login:  "testuser",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer: validIssuer,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid secret key",
			fields: fields{
				secretKey: invalidKey,
				issuer:    validIssuer,
			},
			args: args{
				tokenString: validToken,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid issuer",
			fields: fields{
				secretKey: validKey,
				issuer:    invalidIssuer,
			},
			args: args{
				tokenString: validToken,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty token",
			fields: fields{
				secretKey: validKey,
				issuer:    validIssuer,
			},
			args: args{
				tokenString: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "malformed token",
			fields: fields{
				secretKey: validKey,
				issuer:    validIssuer,
			},
			args: args{
				tokenString: "malformed.token.string",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &services.JWTService{
				SecretKey: tt.fields.secretKey,
				Issuer:    tt.fields.issuer,
			}
			got, err := s.ValidateToken(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Проверяем основные поля, так как временные метки будут отличаться
				if got.UserID != tt.want.UserID || got.Login != tt.want.Login || got.Issuer != tt.want.Issuer {
					t.Errorf("ValidateToken() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
