package main

import (
	"encoding/json"
	"fmt"
	"github.com/caos/oidc/pkg/client/rs"
	"github.com/caos/oidc/pkg/oidc"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	publicURL = "/public"
	protectedURL = "/protected"
	protectedClaimURL = "/protected/{claim}/{value}"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatalln(".env no such file")
	}
	keyPath := os.Getenv("KEY" )
	port := "3001"//os.Getenv("PORT")
	issuer := os.Getenv("ISSUER")

	provider ,err := rs.NewResourceServerFromKeyFile(issuer,keyPath)
	if err != nil {
		logrus.Fatalf("error creating provider %s",err.Error())
	}

	e := echo.New()

	e.GET(publicURL, func(c echo.Context) error {
		return c.String(http.StatusOK,"OK"+ time.Now().String())
	})

	e.GET(protectedURL, func(c echo.Context) error {
		token,ok :=checkToken(c)
		if !ok {
			return nil
		}
		resp,err := rs.Introspect(c.Request().Context(),provider,token)
		if err != nil {
			return c.String(http.StatusForbidden,err.Error())
		}
		data,err := json.Marshal(resp)
		if err!=nil {
			return c.String(http.StatusInternalServerError,err.Error())
		}
		c.Response().Write(data)
		return nil
	})

	e.GET(protectedClaimURL, func(c echo.Context) error {
		token,ok := checkToken(c)
		if !ok {
			return fmt.Errorf("false")
		}
		resp,err := rs.Introspect(c.Request().Context(),provider,token)
		if err != nil {
			return c.String(http.StatusForbidden,err.Error())
		}
		requestedClaim := c.Param("Claim")
		requestedValue :=c.Param("value")
		value,ok := resp.GetClaim(requestedClaim).(string)
		if !ok || value == "" || value != requestedValue {
			return c.String(http.StatusForbidden,"claim does not match")
		}
		_, err = c.Response().Write([]byte("authorized with value " + value))
		if err != nil {
			return err
		}
		return nil
	})


	log.Fatal(e.Start(port))
}

func checkToken(c echo.Context) (string, bool) {
	auth := c.Request().Header.Get("authorization")
	if auth == "" {
		fmt.Errorf("auth header missing %s",http.StatusUnauthorized)
		return "",false
	}
	if !strings.HasPrefix(auth,oidc.PrefixBearer) {
		fmt.Errorf("invalid header %v",http.StatusUnauthorized)
		return "", false
	}
	return strings.TrimPrefix(auth, oidc.PrefixBearer), true
}

