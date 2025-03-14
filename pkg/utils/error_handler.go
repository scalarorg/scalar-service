package utils

import (
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ErrResponse struct {
	Message  string      `json:"message"`
	Metadata interface{} `json:"metadata,omitempty"`
}

func HttpErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if m, ok := err.(*ValidationError); ok {
		c.JSON(http.StatusBadRequest, m)
	} else if m, ok := err.(*echo.HTTPError); ok {
		switch mType := m.Message.(type) {
		case string:
			c.JSON(m.Code, ErrResponse{Message: mType})
		case error:
			c.JSON(m.Code, ErrResponse{Message: mType.Error()})
		}
	} else {
		log.Error().Str("err", err.Error()).Msg("http error")
		if err == pgx.ErrNoRows || err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, ErrResponse{
				Message: "Resource not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: http.StatusText(http.StatusInternalServerError),
		})
	}
}
