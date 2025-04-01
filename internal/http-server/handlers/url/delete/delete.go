package delete

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url_shortener/internal/lib/api/response"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

type URLDelete interface {
	DeleteURL(urlToDelete string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLDelete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body", sl.Err(err))

			render.JSON(w, r, response.Error("failed to decode request body"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("Failed to validate request", sl.Err(err))

			render.JSON(w, r, response.Error("failed to validate request"))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		alias := req.Alias

		id, err := urlSaver.DeleteURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("URL already delete", slog.String("url", req.URL))

			render.JSON(w, r, response.Error("Url already delete"))

			return
		}
		log.Info("Url delete", slog.Int64("url", id))

		responceOk(w, r, alias)
	}
}

func responceOk(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Alias:    alias,
	})
}
