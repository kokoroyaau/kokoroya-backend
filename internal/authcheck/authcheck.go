package authcheck

import (
	"context"
	"errors"

	"kokoroya-backend/internal/jwtauth"
	"kokoroya-backend/internal/session"
)

var ErrInvalid = errors.New("invalid or expired token")

func Verify(ctx context.Context, jwtManager *jwtauth.Manager, sessionManager *session.Manager, token string) (userID int64, role string, jti string, err error) {
	userID, role, jti, err = jwtManager.Parse(token)
	if err != nil {
		return 0, "", "", ErrInvalid
	}

	valid, err := sessionManager.IsValid(ctx, jti)
	if err != nil || !valid {
		return 0, "", "", ErrInvalid
	}

	return userID, role, jti, nil
}
