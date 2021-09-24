package main

import (
	"encoding/json"
	"fmt"
	"github.com/caos/oidc/pkg/client/rp"
	"github.com/caos/oidc/pkg/oidc"
	"github.com/caos/oidc/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	callbackPath = "/auth/callback"
	key = []byte("test1234test1234")
)

func main() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	keyPath := os.Getenv("KEY_PATH")
	issuer := os.Getenv("ISSUER")
	port := "3001"//os.Getenv("PORT")
	scopes := strings.Split(os.Getenv("SCOPES"), " ")

	redirectURI := fmt.Sprintf("http://localhost:%v%v", port, callbackPath)
	cookieHandler :=utils.NewCookieHandler(key,key,utils.WithUnsecure())

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5*time.Second)),
	}
	if clientSecret=="" {
		options = append(options,rp.WithPKCE(cookieHandler))
	}
	if keyPath != "" {
		options = append(options,rp.WithClientKey(keyPath))
	}

	provider ,err := rp.NewRelyingPartyOIDC(issuer,clientID,clientSecret,redirectURI,scopes,options...)
	if err != nil {
		logrus.Fatalf("error creating provider %s",err.Error())
	}

	state := func() string {
		return uuid.New().String()
	}

	http.Handle("/login",rp.AuthURLHandler(state,provider))

	marshalUserinfo := func(w http.ResponseWriter,r *http.Request,tokens *oidc.Tokens, state string , rp rp.RelyingParty,info oidc.UserInfo) {
		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}

	http.Handle(callbackPath,rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo),provider))

	lis := fmt.Sprintf("127.0.0.1:%s", port)
	logrus.Infof("listening on http://%s/", lis)
	logrus.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))


}
