package grpcUtils

import (
	jwtUtils "dormitory-helper-service/internal/utils/jwt"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidateTokenAndGetUserID валидирует JWT токен и возвращает user_id
func ValidateTokenAndGetUserID(token string, jwtSecret []byte) (int, error) {
	if token == "" {
		return 0, status.Errorf(codes.Unauthenticated, "token is required")
	}

	claims, err := jwtUtils.ValidateToken(token, jwtSecret)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	if claims.UserID <= 0 {
		return 0, status.Errorf(codes.Unauthenticated, "invalid user_id in token")
	}

	return claims.UserID, nil
}

// ValidateTokenAndGetUserInfo валидирует JWT токен и возвращает user_id и username
func ValidateTokenAndGetUserInfo(token string, jwtSecret []byte) (userID int, username string, err error) {
	if token == "" {
		return 0, "", fmt.Errorf("token is required")
	}

	claims, err := jwtUtils.ValidateToken(token, jwtSecret)
	if err != nil {
		return 0, "", fmt.Errorf("invalid token: %w", err)
	}

	if claims.UserID <= 0 {
		return 0, "", fmt.Errorf("invalid user_id in token")
	}

	return claims.UserID, claims.Username, nil
}
