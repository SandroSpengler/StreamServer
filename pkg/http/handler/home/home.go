package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/sandrospengler/streamserver/pkg/http/handler/render"
	"github.com/sandrospengler/streamserver/pkg/pages/home"
)

type HomeHandler struct{}

func (h HomeHandler) HandleHomeShow(c echo.Context) error {

	return render.Render(c, home.Home())
}
