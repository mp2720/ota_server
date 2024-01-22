package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"reflect"
)

type TokenService struct {
	cfg *Config
}

type TokenSubject struct {
	name    string
	isBoard bool
}

func (svc *TokenService) New(sub *TokenSubject) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss":   svc.cfg.jwtIssuer,
			"sub":   sub.name,
			"board": sub.isBoard,
		})
	return t.SignedString([]byte(svc.cfg.jwtSigningKey))
}

type InvalidTokenClaimsError struct {
	claim string
}

func (e InvalidTokenClaimsError) Error() string {
	return fmt.Sprintf(
		"Required claim '%s' is missing or invalid in given token",
		e.claim,
	)
}

func verifyClaimType(claims jwt.MapClaims, key string, typeStr string) error {
	if _, ok := claims[key]; !ok {
		return InvalidTokenClaimsError{key}
	}
	if reflect.TypeOf(claims[key]).String() != typeStr {
		return InvalidTokenClaimsError{key}
	}
	return nil
}

func (svc *TokenService) ParseToken(tokenStr string) (*TokenSubject, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(svc.cfg.jwtSigningKey), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if err = verifyClaimType(claims, "iss", "string"); err != nil {
		return nil, err
	}
	if err = verifyClaimType(claims, "sub", "string"); err != nil {
		return nil, err
	}
	if err = verifyClaimType(claims, "board", "bool"); err != nil {
		return nil, err
	}

	return &TokenSubject{
		name:    claims["sub"].(string),
		isBoard: claims["board"].(bool),
	}, nil
}
