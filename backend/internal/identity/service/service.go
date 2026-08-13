package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gmorini/inge-soft-3/backend/internal/identity/dao"
	identityerrors "github.com/gmorini/inge-soft-3/backend/internal/identity/errors"
	"github.com/gmorini/inge-soft-3/backend/internal/identity/repository"
)

type RegisterInput struct {
	FirstName    string
	LastName     string
	Phone        string
	Street       string
	StreetNumber string
	Apartment    string
	City         string
	Province     string
	Username     string
	Email        string
	Password     string
}

type RegisteredUser struct {
	ID       int64
	Username string
	Email    string
}

type LoginInput struct {
	Username string
	Password string
}

type LoginResult struct {
	AccessToken string
	ExpiresIn   int
}

type Service struct {
	repository *repository.Repository
	tokens     *TokenManager
	dummyHash  string
}

func New(repository *repository.Repository, tokens *TokenManager) (*Service, error) {
	dummyHash, err := hashPassword("dummy-password-never-used!123")
	if err != nil {
		return nil, fmt.Errorf("create dummy credential hash: %w", err)
	}
	return &Service{repository: repository, tokens: tokens, dummyHash: dummyHash}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	fields := make(map[string][]string)
	if input.Username == "" {
		fields["username"] = []string{"Es obligatorio."}
	}
	if input.Password == "" {
		fields["password"] = []string{"Es obligatoria."}
	}
	if len(fields) > 0 {
		return LoginResult{}, &identityerrors.ValidationError{Fields: fields}
	}

	credentials, err := s.repository.FindCredentialsByUsername(
		ctx,
		strings.ToLower(input.Username),
	)
	if errors.Is(err, identityerrors.ErrUserNotFound) {
		if _, verifyErr := verifyPassword(input.Password, s.dummyHash); verifyErr != nil {
			return LoginResult{}, fmt.Errorf("verify dummy credential: %w", verifyErr)
		}
		return LoginResult{}, identityerrors.ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("load login credentials: %w", err)
	}

	valid, err := verifyPassword(input.Password, credentials.PasswordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("verify login password: %w", err)
	}
	if !valid {
		return LoginResult{}, identityerrors.ErrInvalidCredentials
	}

	accessToken, err := s.tokens.Issue(credentials.ID, credentials.Username)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue login token: %w", err)
	}
	return LoginResult{AccessToken: accessToken, ExpiresIn: int(tokenLifetime.Seconds())}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (RegisteredUser, error) {
	normalized, err := normalizeAndValidateRegistration(input)
	if err != nil {
		return RegisteredUser{}, err
	}

	passwordHash, err := hashPassword(normalized.Password)
	if err != nil {
		return RegisteredUser{}, fmt.Errorf("hash registration password: %w", err)
	}

	created, err := s.repository.CreateUser(ctx, dao.CreateUserParams{
		FirstName:    normalized.FirstName,
		LastName:     normalized.LastName,
		Phone:        normalized.Phone,
		Street:       normalized.Street,
		StreetNumber: normalized.StreetNumber,
		Apartment:    normalized.Apartment,
		City:         normalized.City,
		Province:     normalized.Province,
		Username:     normalized.Username,
		Email:        normalized.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return RegisteredUser{}, fmt.Errorf("create registered user: %w", err)
	}

	return RegisteredUser{
		ID:       created.ID,
		Username: created.Username,
		Email:    created.Email,
	}, nil
}
