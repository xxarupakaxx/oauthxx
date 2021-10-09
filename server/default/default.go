package Example

import (
	"context"
	"crypto/sha256"
	"github.com/caos/oidc/pkg/op"
	"github.com/gorilla/mux"
	"github.com/xxarupakaxx/oauthxx/internal/mock"
	"log"
	"net/http"
	"text/template"
)

func defaultExample() {
	ctx := context.Background()
	port:="9998"
	config :=&op.Config{
		Issuer: "http://localhost:9998/",
		CryptoKey: sha256.Sum256([]byte("test")),
	}

	storage := mock.NewAuthStorage()
	handler ,err := op.NewOpenIDProvider(ctx,config,storage,op.WithCustomTokenEndpoint(op.NewEndpoint("test")))
	if err != nil {
		log.Fatalln(err)
	}
	router := handler.HttpHandler().(*mux.Router)
	router.Methods("GET").Path("/login").HandlerFunc(HandleLogin)
	router.Methods("GET").Path("/login").HandlerFunc(HandleCallback)
	server := &http.Server{
		Addr: ":"+port,
		Handler: router,
	}
	err =server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
	<-ctx.Done()
}

func HandleLogin(w http.ResponseWriter, _ *http.Request) {
	tpl :=  `
	<!DOCTYPE html>
	<html>
		<head>
			<meta charset="UTF-8">
			<title>Login</title>
		</head>
		<body>
			<form method="POST" action="/login">
				<input name="client"/>
				<button type="submit">Login</button>
			</form>
		</body>
	</html>`

	t,err := template.New("login").Parse(tpl)
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
	}
	err = t.Execute(w,nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		return
	}
	client := r.FormValue("client")
	http.Redirect(w, r, "/authorize/callback?id="+client, http.StatusFound)
}
