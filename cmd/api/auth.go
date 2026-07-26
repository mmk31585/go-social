package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"social/internal/store"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type registerUserPayload struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=32"`
}
type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// RegisterUser godoc
//
//	@Summary		Registers a new user
//	@Description	Registers a new user with an invitation token sent via email
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			body	body		registerUserPayload	true	"User registration payload"
//	@Success		201		{object}	store.User
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload registerUserPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role: store.Role{
			Name: "user",
		},
	}
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	plaintToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plaintToken))
	hasToken := hex.EncodeToString(hash[:])

	ctx := context.Background()
	if err := app.store.Users.CreateAndInvite(ctx, user, hasToken, app.config.mail.exp); err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestError(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	userWithToken := &UserWithToken{
		User:  user,
		Token: plaintToken,
	}
	// activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plaintToken)
	// isProdEnv := app.config.env == "production"
	// vars := struct {
	// 	UserName      string
	// 	ActivationURL string
	// }{
	// 	UserName:      user.Username,
	// 	ActivationURL: activationURL,
	// }
	// err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	// if err != nil {
	// 	app.logger.Errorw("error sending welcom email", "error", err)
	// 	if err := app.store.Users.Delete(ctx, user.ID); err != nil {
	// 		app.logger.Errorw("error deleting user", "error", err)
	// 	}
	// 	app.internalServerError(w, r, err)
	// 	return
	// }

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// GenerateTokem=n godoc
//
//	@Summary		generate token
//	@Description	generate token
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateUserTokenPayload	true	"Create user token"
//	@Success		200		{object}	string
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserTokenPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	log.Print(user)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedErrorResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.iss,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}
}
