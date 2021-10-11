package main

import (
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
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
	e.Static("/static","template/client")

	e.GET("/",indexClient)
	e.GET("/authorize",authorizeClient)
	e.GET("/callback",callbackClient)
	e.GET("/fetch_resource",resourceClient)


	e.Logger.Fatal(e.Start(":9000"))
}

func resourceClient(c echo.Context) error {
	return c.Render(http.StatusOK,"data",nil)

}

func callbackClient(c echo.Context) error {
	return c.Render(http.StatusOK,"index",nil)
}

func authorizeClient(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently,"https://twitter.com/home")
}

func indexClient(c echo.Context) error {
	return c.Render(http.StatusOK,"index",nil)
}






