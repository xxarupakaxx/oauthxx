package main

import (
	"fmt"
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
	t:=&ResourceTemplate{templates: template.Must(template.ParseGlob("template/resource/*.html"))}

	e:=echo.New()
	e.Static("/static","template/resource")
	e.Renderer = t
	e.GET("/", indexResource)
	e.POST("/resource",postAccessToken)

	e.Logger.Debug(e.Start(":9002"))
}

func indexResource(c echo.Context) error {
	return c.Render(http.StatusOK,"index",nil )
}



func postAccessToken(c echo.Context) error {
	auth := c.Request().Header.Get("authorization")
	fmt.Println(auth)
	var inToken string
	if auth != "" {
		inToken = auth[len("bearer "):]
	}

	fmt.Println("Incoming token: ",inToken)

	var resource  = struct {
		Name string `json:"name"`
		Description string `json:"description"`
	}{
		Name:        "Protected Resource",
		Description: "This data has been protected by OAuth 2.0",
	}

	return c.JSON(http.StatusOK,resource)
}
