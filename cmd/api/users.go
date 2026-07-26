package main

import (
	"context"
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// ActivateUser godoc
//
//	@Summary		Activate user by token
//	@Description	Activates a user account using a one-time verification token.
//	@Tags			users
//	@Param			token	path	string	true	"Activation token"
//	@Success		204		"User successfully activated (no content)"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, "user is active"); err != nil {
		app.internalServerError(w, r, err)
	}

}

type userKey string

const userCtx userKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		Bearer
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := r.Context()

	user, err := app.getUser(ctx, userID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) handleUserAction(
	w http.ResponseWriter,
	r *http.Request,
	action func(ctx context.Context, followerID, userID int64) error,
) {
	follower := GetUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := context.Background()
	if err := action(ctx, follower.ID, followedID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by the authenticated user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"User ID to follow"
//	@Param			body	body		followerPayload	true	"Follower payload"
//	@Success		200		{object}	nil
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		Bearer
//	@Router			/users/{id}/follow [patch]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	app.handleUserAction(w, r, app.store.Followers.Follow)
}

// UnfollowUser godoc
//
//	@Summary		Unfollows a user
//	@Description	Unfollows a user by the authenticated user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"User ID to unfollow"
//	@Param			body	body		followerPayload	true	"Follower payload"
//	@Success		200		{object}	nil
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		Bearer
//	@Router			/users/{id}/unfollow [patch]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	app.handleUserAction(w, r, app.store.Followers.Unfollow)
}
func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}
		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, id)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
				return
			}
		}
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func GetUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
