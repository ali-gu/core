package auth

import "context"

type SignUpResult struct {
	ID string
}

type AuthenticatedUser struct {
	ID    string
	Email string
}

type Session struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	ExpiresAt    int64
}

type SignInResult struct {
	AuthenticatedUser
	Session
}

type InviteResult struct {
	ID string
}

type IAuth interface {
	SignUp(ctx context.Context, username string, password string) (*SignUpResult, error)

	SignIn(ctx context.Context, email string, password string) (*SignInResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*SignInResult, error)
	Authenticate(ctx context.Context, accessToken string) (*AuthenticatedUser, error)
	SignOut(ctx context.Context, accessToken string) error

	Invite(ctx context.Context, email string) (*InviteResult, error)
}
