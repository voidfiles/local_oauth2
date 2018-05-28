package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sync"

	oidc "github.com/coreos/go-oidc"
)

var clientID = flag.String("clientid", "", "The oauth2 client id")
var host = flag.String("host", "", "The oauth2 host")
var port = flag.String("port", "10000", "The port to run the server at")
var verifier *oidc.IDTokenVerifier

var nonce string

func randString() string {
	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	return base64.URLEncoding.EncodeToString(b)
}

func buildAuthorizeURL() string {
	u, err := url.Parse(fmt.Sprintf("https://%s/authorize", *host))
	if err != nil {
		log.Fatal(err)
	}
	nonce = randString()
	q := u.Query()
	q.Set("response_type", "id_token")
	q.Set("client_id", *clientID)
	q.Set("scope", "openid profile email")
	q.Set("redirect_uri", "http://localhost:10000/")
	q.Set("nonce", nonce)
	u.RawQuery = q.Encode()
	return fmt.Sprint(u)
}

func viewIndex(w http.ResponseWriter, r *http.Request) {
	message := `
	<script>
	var myRequest = new Request('/save?' + window.location.hash.slice(1));
	fetch(myRequest).then(function(response) {
	  window.close();
	});
	</script>
	`
	w.Write([]byte(message))
}

func saveToken(w http.ResponseWriter, r *http.Request) {
	idTokenRaw := r.URL.Query().Get("id_token")
	ctx := context.TODO()
	idToken, err := verifier.Verify(ctx, idTokenRaw)
	if err != nil {
		log.Fatal(err)
	}

	if nonce != idToken.Nonce {
		log.Fatalf("Nonce does not match. Before: %s From Request: %s", nonce, r.URL.Query().Get("nonce"))
		w.Write([]byte("Nonce mismatch"))
		return
	}
	nonce = ""

	fmt.Fprintf(os.Stderr, "id_token: %s\n", idToken)
	fmt.Printf("%s\n", idTokenRaw)
	w.Write([]byte(""))
	os.Exit(0)
}

func main() {
	flag.Parse()
	ctx := context.TODO()
	provider, err := oidc.NewProvider(ctx, fmt.Sprintf("https://%s/", *host))
	if err != nil {
		log.Fatalf("Failed to create a provider: %s", err)
	}
	verifier = provider.Verifier(&oidc.Config{ClientID: *clientID})
	http.HandleFunc("/", viewIndex)
	http.HandleFunc("/save", saveToken)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil); err != nil {
			panic(err)
		}
	}()
	urlToOpen := buildAuthorizeURL()
	fmt.Fprintf(os.Stderr, "Open URL: %s\n", urlToOpen)
	exec.Command("open", buildAuthorizeURL()).Start()
	wg.Wait()
}
