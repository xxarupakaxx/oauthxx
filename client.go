package main

import (
	"github.com/labstack/echo/v4"
	"io"
	"text/template"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func main() {
	t:=&Template{templates: template.Must(template.ParseGlob("template/client/*.html"))}
	e :=echo.New()

	e.Renderer = t


	e.Logger.Fatal(e.Start(":9000"))
}





