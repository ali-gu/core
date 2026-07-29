package supabase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ali-gulzar/speechory-core/internal/services/auth"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/supabase-community/gotrue-go/types"
	supa "github.com/supabase-community/supabase-go"
)

type Supabase struct {
	client *supa.Client
}

func NewSupabase(client *supa.Client) *Supabase {
	return &Supabase{
		client: client,
	}
}

var _ auth.IAuth = (*Supabase)(nil)

func (s *Supabase) SignUp(_ context.Context, username string, password string) (*auth.SignUpResult, error) {
	resp, err := s.client.Auth.Signup(types.SignupRequest{
		Email:    username,
		Password: password,
	})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("supabase: sign up: %w", err))
	}

	return &auth.SignUpResult{ID: resp.User.ID.String()}, nil
}

func (s *Supabase) SignIn(_ context.Context, email string, password string) (*auth.SignInResult, error) {
	resp, err := s.client.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, rerror.New(errors.New("invalid credentials")).WithKind(rerror.Permission)
	}

	return &auth.SignInResult{
		AuthenticatedUser: auth.AuthenticatedUser{
			ID:    resp.User.ID.String(),
			Email: resp.User.Email,
		},
		Session: auth.Session{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			TokenType:    resp.TokenType,
			ExpiresIn:    resp.ExpiresIn,
			ExpiresAt:    resp.ExpiresAt,
		},
	}, nil
}

func (s *Supabase) RefreshToken(_ context.Context, refreshToken string) (*auth.SignInResult, error) {
	resp, err := s.client.Auth.RefreshToken(refreshToken)
	if err != nil {
		return nil, rerror.New(errors.New("invalid refresh token")).WithKind(rerror.Permission)
	}

	return &auth.SignInResult{
		AuthenticatedUser: auth.AuthenticatedUser{
			ID:    resp.User.ID.String(),
			Email: resp.User.Email,
		},
		Session: auth.Session{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			TokenType:    resp.TokenType,
			ExpiresIn:    resp.ExpiresIn,
			ExpiresAt:    resp.ExpiresAt,
		},
	}, nil
}

func (s *Supabase) SignOut(_ context.Context, accessToken string) error {
	if err := s.client.Auth.WithToken(accessToken).Logout(); err != nil {
		return rerror.New(fmt.Errorf("supabase: sign out: %w", err))
	}
	return nil
}

func (s *Supabase) Authenticate(_ context.Context, accessToken string) (*auth.AuthenticatedUser, error) {
	resp, err := s.client.Auth.WithToken(accessToken).GetUser()
	if err != nil {
		return nil, rerror.New(errors.New("invalid token")).WithKind(rerror.Permission)
	}

	return &auth.AuthenticatedUser{
		ID:    resp.User.ID.String(),
		Email: resp.User.Email,
	}, nil
}

func (s *Supabase) Invite(_ context.Context, email string) (*auth.InviteResult, error) {
	resp, err := s.client.Auth.Invite(types.InviteRequest{Email: email})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("supabase: invite: %w", err))
	}

	return &auth.InviteResult{ID: resp.User.ID.String()}, nil
}
