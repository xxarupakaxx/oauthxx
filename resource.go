package main

import (
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"text/template"
)
type ResourceTemplate struct {
	templates *template.Template
}
func (t *ResourceTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func main() {
	t:=&Template{templates: template.Must(template.ParseGlob("template/resource/*.html"))}

	e:=echo.New()

	e.Renderer = t
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK,"good")
	})

	e.Logger.Debug(e.Start(":9002"))
}