package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strings"
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

	e.Logger.Debug(e.Start(":9002"))
}

func indexResource(c echo.Context) error {
	return c.Render(http.StatusOK,"index",nil )
}

var resource  = struct {
	name string `json:"name"`
	description string `json:"description"`
}{
	name:        "Protected Resource",
	description: "This data has been protected by OAuth 2.0",
}

func getAccessToken(c echo.Context) error {
	auth := c.Request().Header.Get("authorization")
	var inToken string
	if auth != "" && strings.Index(auth, "bearer") == 0 {
		inToken = auth[len("bearer "):]
	}

	fmt.Println("Incoming token: ",inToken)
	cli := DBconnect()
	collection := cli.Database("grpc").Collection("test")

}