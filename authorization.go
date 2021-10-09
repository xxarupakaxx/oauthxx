package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
	"unsafe"
)

type AuthTemplate struct {
	templates *template.Template
}
type Errors struct {
	Error string
}
type Info struct {
	Client        string
	ClientSecret  string
	Scope         []string
	RedirectURI   string
	AuthEndpoint  string
	TokenEndpoint string
	Name          string
	ClientLogoURI string
	ClientURI     string
}
type ClientInfo struct {
	Client        string
	ClientSecret  string
	Name          string
	Scope         []string
	ClientLogoURI string
	ClientURI     string
	RedirectURI   string
	ReqID string
	URL *url.URL
}

var CI = ClientInfo{
	Client:        "oauth-client-1",
	ClientSecret:  "oauth-client-secret-1",
	Name:          "xxarupakaxx",
	Scope:         []string{"foo", "bar"},
	ClientLogoURI: "https://user-images.githubusercontent.com/67729473/120451954-d50bd800-c3cc-11eb-92dd-84e20cbd323c.png",
	ClientURI:     "http://localhost:9000",
	RedirectURI:   "http://localhost:9000/callback",
}

type ApproveInfo struct {
	AuthorizationEndpointRequest url.Values
	Scope                        []string
}
var codes map[string]ApproveInfo

func (t *AuthTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func main() {
	t := &AuthTemplate{templates: template.Must(template.ParseGlob("template/auth/*.html"))}

	e := echo.New()

	e.Renderer = t

	e.GET("/authorize", authorize)
	e.GET("/", index)
	e.GET("/error", errorHandler)
	e.POST("/approve",approve)

	e.Logger.Fatal(e.Start(":9001"))

}

func errorHandler(c echo.Context) error {

	var e Errors
	e.Error = "Invalid"
	return c.Render(http.StatusInternalServerError, "error", e)
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

	return c.Render(http.StatusOK, "index", info)
}

func authorize(c echo.Context) error {
	var reqid string
	uri, err := url.ParseRequestURI(c.Request().RequestURI)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	q := uri.Query()
	client_id := q.Get("client_id")
	if client_id != CI.Client {
		e := Errors{fmt.Sprintf("Unknown client %s,%s", CI.Client, client_id)}
		return c.Render(http.StatusBadRequest, "error", e)
	} else if !contains(q["redirect_uri"], CI.RedirectURI) {
		e := Errors{fmt.Sprintf("Mismatched redirect URI, expected %s ", CI.RedirectURI)}
		return c.Render(http.StatusBadRequest, "error", e)
	} else {
		rscope := strings.Join(q["scope"], ",")
		cscope := strings.Join(CI.Scope, ",")
		if !strings.Contains(rscope, cscope) {
			urlParsed, err := url.Parse(q.Get("redirect_uri"))
			if err != nil {
				e := Errors{fmt.Sprintf("Mismatched redirect URI, expected %s ", CI.RedirectURI)}
				return c.Render(http.StatusBadRequest, "error", e)
			}
			c.Redirect(http.StatusOK, urlParsed.String())
		}
		reqid, err = MakeRandomStr(8)
		if err != nil {
			return c.Render(http.StatusBadRequest, "error", Errors{err.Error()})
		}
	}
	CI.URL = uri
	fmt.Println(reqid)
	return c.Render(http.StatusOK, "approve", CI)
}

func approve(c echo.Context) error {
	query := CI.URL.Query()
	if query.Get("response_type") == "code" {

		code, err := MakeRandomStr(8)
		if err != nil {
			return c.Render(http.StatusInternalServerError,"error",Errors{err.Error()})
		}
		rscope := strings.Join(query["scope"], ",")
		cscope := strings.Join(CI.Scope, ",")
		if !strings.Contains(rscope,cscope) {
			urlParsed := query.Get("redirect_uri")
			urlParsed +="?state="
			urlParsed += query.Get("state")
			fmt.Println(urlParsed)
			if err != nil {
				e := Errors{fmt.Sprintf("Mismatched redirect URI, expected %s ", CI.RedirectURI)}
				return c.Render(http.StatusBadRequest, "error", e)
			}
			return  c.Redirect(http.StatusMovedPermanently, urlParsed)
		}
		codes[code] =ApproveInfo{
			AuthorizationEndpointRequest: query,
			Scope:                        query["scope"],
		}
		urlParsed := query.Get("redirect_uri")
		if err != nil {
			e := Errors{fmt.Sprintf("Mismatched redirect URI, expected %s ", CI.RedirectURI)}
			return c.Render(http.StatusBadRequest, "error", e)
		}
		return  c.Redirect(http.StatusMovedPermanently, urlParsed)
	}
	return c.Redirect(http.StatusMovedPermanently,"http://localhost:9000")
}

func Token(c echo.Context) error {
	auth := c.Request().Header.Get("authorization")
	dec,err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
	if err!=nil {
		log.Println("Couldnot decode auth:%w",err)
	}
	clientCredentials := *(*string)(unsafe.Pointer(&dec))
	clientInfo := strings.Split(clientCredentials,":")
	if clientInfo[0] != CI.Client {
		e := Errors{fmt.Sprintf("Unknown client %s,%s", CI.Client, clientInfo[0])}
		return c.Render(http.StatusUnauthorized, "error", e)
	}
	if clientInfo[1] != CI.ClientSecret {
		e := Errors{fmt.Sprintf("Mismatched client secret, expected %s got %s", CI.ClientSecret, clientInfo[1])}
		return c.Render(http.StatusUnauthorized, "error", e)
	}
	if c.Request().URL.Query().Get("grant_type") == "authorization_code"{
		code := codes[ c.Request().URL.Query().Get("code")]
		if code.AuthorizationEndpointRequest != nil {
			accessToken,err := MakeRandomStr(16)
			if err != nil {
				log.Println(err.Error())
			}
			csope := strings.Join(code.Scope," ")
			cli := DBconnect()
			collection := cli.Database("grpc").Collection("test")
			_, err = collection.InsertOne(context.Background(), struct {
				accessToken string `bson:"access_token"`
				clientId string `bson:"client_id"`
				scope string `bson:"scope"`
			}{
				accessToken: accessToken,
				clientId:    clientInfo[0],
				scope:       csope,
			})
			if err != nil {
				return c.String(http.StatusInternalServerError,err.Error())
			}



		}
	}

}
func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}
func MakeRandomStr(digit uint32) (string, error) {
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
func DBconnect() *mongo.Client {
	cli,err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil { log.Fatal(err) }
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = cli.Connect(ctx)
	if err != nil { log.Fatal(err) }
	defer cli.Disconnect(ctx)

	return cli
}