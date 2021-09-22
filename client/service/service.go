package main

import (
	"context"
	"encoding/json"
	"github.com/caos/oidc/pkg/client/profile"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
)

var client *http.Client = http.DefaultClient

func main() {
	keyPath := os.Getenv("KEY_PATH")
	issuer := os.Getenv("ISSUER")
	port := os.Getenv("PORT")
	scopes := strings.Split(os.Getenv("SCOPES")," ")

	if keyPath != "" {
		ts ,err := profile.NewJWTProfileTokenSourceFromKeyFile(issuer,keyPath,scopes)
		if err != nil {
			logrus.Fatalf("error creating token source %s",err.Error())
		}
		client = oauth2.NewClient(context.Background(),ts)


	}

	http.HandleFunc("/jwt-profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method=="GET" {
			tpl:=`
	<!DOCTYPE html>
	<html>
		<head>
			<meta charset="UTF-8">
			<title>Login</title>
		</head>
		<body>
			<form method="POST" action="/jwt-profile" enctype="multipart/form-data">
				<label for="key">Select a key file:</label>
				<input type="file" accept=".json" id="key" name="key">
				<button type="submit">Get Token</button>
			</form>
		</body>
	</html>`
			t,err := template.New("login").Parse(tpl)
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			err = t.Execute(w,nil)
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
		}else {
			err := r.ParseMultipartForm(4 << 10)
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			file,_,err:=r.FormFile("key")
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			defer file.Close()

			key,err := ioutil.ReadAll(file)
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			ts,err := profile.NewJWTProfileTokenSourceFromKeyFileData(issuer,key,scopes)
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			client = oauth2.NewClient(context.Background(),ts)
			token,err:=ts.Token()
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			data,err := json.Marshal(token)
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			w.Write(data)
		}
	})
}