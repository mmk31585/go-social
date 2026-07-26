package main

import (
	"net/http"
	"social/internal/store"
)

// GetUserFeed godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the paginated feed for the authenticated user
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"		default(20)
//	@Param			offset	query		int		false	"Offset"	default(0)
//	@Param			sort	query		string	false	"Sort"		Enums(asc, desc)
//	@Param			tags	query		string	false	"Tags (comma separated)"
//	@Param			search	query		string	false	"Search term"
//	@Param			since	query		string	false	"Since (RFC3339 or YYYY-MM-DD)"
//	@Param			until	query		string	false	"Until (RFC3339 or YYYY-MM-DD)"
//	@Success		200		{array}		store.PostWithMetaData
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		Bearer
//	@Router			/posts/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  30,
		Offset: 0,
		Sort:   "desc",
	}
	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := r.Context()
	feed, err := app.store.Posts.GetUserFeed(ctx, int64(109), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
