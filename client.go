package main

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"math/rand"
	"net/http"
	url2 "net/url"
	"text/template"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
type ClientErrors struct {
	Error string
}
var accessToken string
var scope string
var state string
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
	queryState := c.FormValue("state")
	fmt.Println("state:",state,"query",queryState)
	if state != queryState {
		return c.Render(http.StatusInternalServerError,"error",ClientErrors{fmt.Sprintf("mismatch state:%s,%s",state,queryState)})
	}
	return c.Render(http.StatusOK,"index",nil)
}

func authorizeClient(c echo.Context) error {
	accessToken = ""
	scope = ""
	state,_ = cliMakeRandomStr(16);

	url,_ := url2.Parse("http://localhost:9001/authorize")
	url.Query().Add("response_type","code")
	url.Query().Add("scope","foo bar")
	url.Query().Add("client_id","oauth-client-1")
	url.Query().Add("redirect_uri","http://localhost:9000/callback")
	url.Query().Add("state",state)

	c.Request().URL = url
	fmt.Println(url.String())
	return c.Redirect(http.StatusFound,url.String())
}


func indexClient(c echo.Context) error {
	data := struct {
		AccessToken string `json:"access_token"`
		Scope string `json:"scope"`
	}{
		AccessToken: accessToken,
		Scope: scope,
	}
	return c.Render(http.StatusOK,"index",data)
}

func cliMakeRandomStr(digit uint32) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 乱数を生成
	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("unexpected error...")
	}

	// letters からランダムに取り出して文字列を生成
	var result string
	for _, v := range b {
		// index が letters の長さに収まるように調整
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}




