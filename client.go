package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	url2 "net/url"
	"strings"
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
	if accessToken =="" {
		return c.Render(http.StatusBadRequest,"error", ClientErrors{Error: "Missing AccessToken"})
	}
	fmt.Printf("making accessToken %s",accessToken)

	req,err := http.NewRequest("POST","http://localhost:9002/resource",nil)
	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Authorization","Bearer "+ accessToken)
	client := &http.Client{}
	res ,err :=client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >=200 && res.StatusCode<300 {

		body,_:= ioutil.ReadAll(res.Body)
		data := struct {
			Name string `json:"name"`
			Description string `json:"description"`
		}{}
		if err:= json.Unmarshal(body,&data);err!=nil{
			return c.Render(http.StatusInternalServerError,"error",nil)
		}
		fmt.Println(data)
		return c.Render(http.StatusOK,"data",data)
	}

	return c.Render(http.StatusOK,"data",nil)
}

func callbackClient(c echo.Context) error {
	queryState := c.FormValue("state")
	fmt.Println("state:",state,"query",queryState)
	if state != queryState {
		return c.Render(http.StatusInternalServerError,"error",nil)
	}
	code := c.FormValue("code")
	url:=url2.Values{}
	url.Set("grant_type","authorization_code")
	url.Set("code",code)
	url.Set("redirect_uri","http://localhost:9000/callback")

	req,err := http.NewRequest("POST","http://localhost:9001/token",strings.NewReader(url.Encode()))
	fmt.Println(req,  "     req")
	if err != nil {
		fmt.Println("request error")
		return c.Render(http.StatusInternalServerError,"error",nil)
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization","Basic "+encodeClientCredentials("oauth-client-1","oauth-client-secret-1"))
	client := &http.Client{}
	res ,err :=client.Do(req)
	if err != nil {
		fmt.Println("response error")
		return c.Render(http.StatusInternalServerError,"error",nil)
	}
	defer res.Body.Close()
	fmt.Println(res.StatusCode)
	if res.StatusCode >=200 && res.StatusCode<300 {
		data := struct {
			AccessToken string `json:"access_token"`
			TokenType string `json:"token_type"`
			Scope string `json:"scope"`
		}{}
		body,_:= ioutil.ReadAll(res.Body)
		if err:= json.Unmarshal(body,&data);err!=nil{
			return c.Render(http.StatusInternalServerError,"error",nil)
		}
		fmt.Println(data)
		accessToken = data.AccessToken
		scope = data.Scope
		return c.Render(http.StatusOK,"index", data)
	}
	return c.Render(http.StatusOK,"index",nil)
}

func authorizeClient(c echo.Context) error {
	accessToken = ""
	scope = ""
	state,_ = cliMakeRandomStr(16);
	url := url2.Values{}
	url.Set("response_type","code")
	url.Set("scope","foo bar")
	url.Set("client_id","oauth-client-1")
	url.Set("redirect_uri","http://localhost:9000/callback")
	url.Set("state",state)

	fmt.Println("http://localhost:9001/authorize"+url.Encode())
	return c.Redirect(http.StatusFound,"http://localhost:9001/authorize?"+url.Encode())
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

func encodeClientCredentials(clientId, clientSecret string) string {
	var a = clientId + ":" + clientSecret
	return base64.StdEncoding.EncodeToString([]byte(a))
}


