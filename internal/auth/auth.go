package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
)

func MakeJwtTokenString(signgingKey []byte, userId, tokenName string, tokenExpiration int) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    userId,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(tokenExpiration) * time.Minute)),
		Subject:   tokenName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signgingKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func UpdateRefreshToken(signgingKey []byte, userId uuid.UUID, expirationTime int, w http.ResponseWriter, DB *database.Queries) error {
	// create new jwt refresh token
	jwtRefreshTokenString, err := MakeJwtTokenString(signgingKey, userId.String(), "refresh_token", expirationTime)
	if err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(time.Duration(expirationTime) * time.Minute)
	//  save to http cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshtoken",
		Value:    jwtRefreshTokenString,
		Expires:  expiresAt,
		HttpOnly: true,
		// Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	// save refresh to db
	err = DB.UpdateRefreshToken(context.Background(), database.UpdateRefreshTokenParams{
		ExpiresAt: expiresAt,
		Token:     jwtRefreshTokenString,
		UserID:    userId,
	})

	if err != nil {
		return err
	}

	return nil
}
func CreateRefreshToken(signgingKey []byte, userId uuid.UUID, expirationTime int, w http.ResponseWriter, DB *database.Queries) error {
	// create new jwt refresh token
	jwtRefreshTokenString, err := MakeJwtTokenString(signgingKey, userId.String(), "refresh_token", expirationTime)
	if err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(time.Duration(expirationTime) * time.Minute)
	//  save to http cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    jwtRefreshTokenString,
		Expires:  expiresAt,
		HttpOnly: true,
		// Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	// save refresh to db
	_, err = DB.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		ExpiresAt: expiresAt,
		Token:     jwtRefreshTokenString,
		UserID:    userId,
	})

	if err != nil {
		return err
	}

	return nil
}
