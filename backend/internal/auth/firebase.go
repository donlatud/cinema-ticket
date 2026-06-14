package auth

import (
	"context"
	"errors"
	"fmt"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var ErrInvalidFirebaseToken = errors.New("invalid firebase id token")

type FirebaseVerifier struct {
	client *fbauth.Client
}

func NewFirebaseVerifier(ctx context.Context, credentialsPath string) (*FirebaseVerifier, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase auth client: %w", err)
	}

	return &FirebaseVerifier{client: client}, nil
}

type FirebaseIdentity struct {
	UID   string
	Email string
	Name  string
}

func (f *FirebaseVerifier) VerifyIDToken(ctx context.Context, idToken string) (*FirebaseIdentity, error) {
	if idToken == "" {
		return nil, ErrInvalidFirebaseToken
	}

	token, err := f.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFirebaseToken, err)
	}

	email, _ := token.Claims["email"].(string)
	name, _ := token.Claims["name"].(string)

	if email == "" {
		return nil, fmt.Errorf("%w: email claim is missing", ErrInvalidFirebaseToken)
	}

	return &FirebaseIdentity{
		UID:   token.UID,
		Email: email,
		Name:  name,
	}, nil
}
