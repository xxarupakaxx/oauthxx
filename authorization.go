package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
	"text/template"
)
type AuthTemplate struct {
	templates *template.Template
}
type Errors struct {
	Error string
}
type Info struct {
	Client string
	ClientSecret string
	Scope []string
	RedirectURI string
	AuthEndpoint string
	TokenEndpoint string
	Name string
	ClientLogoURI string
	ClientURI string
}
type ClientInfo struct {
	Client string
	ClientSecret string
	Name string
	Scope []string
	ClientLogoURI string
	ClientURI string
	RedirectURI string
}

var CI = ClientInfo{
	Client:        "oauth-client-1",
	ClientSecret:  "oauth-client-secret-1",
	Name:          "xxarupakaxx",
	Scope:         []string{"foo", "bar"},
	ClientLogoURI: "https://user-images.githubusercontent.com/67729473/120451954-d50bd800-c3cc-11eb-92dd-84e20cbd323c.png",
	ClientURI:     "http://localhost:9000",
	RedirectURI:     "http://localhost:9000/callback",
}

func (t *AuthTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func main() {
	t := &AuthTemplate{templates: template.Must(template.ParseGlob("template/auth/*.html"))}

	e := echo.New()

	e.Renderer = t

	e.GET("/authorize",authorize)
	e.GET("/",index)
	e.GET("/error",errorHandler)

	e.Logger.Fatal(e.Start(":9001"))

}

func errorHandler(c echo.Context) error {
	
	var e Errors
	e.Error = "Invalid"
	return c.Render(http.StatusInternalServerError,"error",e)
}
func index(c echo.Context) error {

	info := Info{
		Client:        CI.Client,
		ClientSecret:  CI.ClientSecret,
		Scope:         []string{"foo", "bar"},
		RedirectURI:   "http://localhost:9000/callback",
		AuthEndpoint:  "http://localhost:9001/authorize",
		TokenEndpoint: "http://localhost:9001/token",
	}

	return c.Render(http.StatusOK,"index",info)
}

func authorize(c echo.Context) error {
	uri,err:=url.ParseRequestURI(c.Request().RequestURI)
	if err!=nil{
		return c.String(http.StatusInternalServerError,err.Error())
	}
	q := uri.Query()
	if contains(q["client_id"],CI.Client) {
		e := Errors{fmt.Sprintf("Unknown client %s",CI.Client)}
		return c.Render(http.StatusBadRequest,"error",e)
	}else if contains(q["redirect_uri"],CI.RedirectURI) {
		e := Errors{fmt.Sprintf("Mismatched redirect URI, expected %s ",CI.ClientURI)}
		return c.Render(http.StatusBadRequest,"error",e)
	}

	return c.Render(http.StatusOK,"approve",CI)
}
func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}