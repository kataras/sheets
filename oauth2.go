package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// ScopeReadOnly is the readonly oauth2 scope.
	ScopeReadOnly = "https://www.googleapis.com/auth/spreadsheets.readonly"
	// ScopeReadWrite is the full-access oauth2 scope.
	ScopeReadWrite = "https://www.googleapis.com/auth/spreadsheets"
)

// ServiceAccount is an oauth2 authentication function which
// can be passed on the `New` package-level function.
//
// It requires Sheet -> Share button to the email of the service account
// but it does not need to keep and maintain a token.
//
// It panics on errors.
func ServiceAccount(ctx context.Context, serviceAccountFile string, scopes ...string) http.RoundTripper {
	b, err := ioutil.ReadFile(serviceAccountFile)
	if err != nil {
		log.Fatalf("Unable to read service account secret file: %v", err)
	}

	if len(scopes) == 0 {
		scopes = []string{ScopeReadOnly}
	}

	config, err := google.JWTConfigFromJSON(b, scopes...)
	client := config.Client(ctx)
	return client.Transport
}

// Token is an oauth2 authentication function which
// can be passed on the `New` package-level function.
// It accepts a token file and optionally scopes (see `ScopeReadOnly` and `ScopeReadWrite` package-level variables).
// At the future it may accept scopes from different APIs (e.g google drive to save the spreadsheets on a specified folder).
//
// It panics on errors.
func Token(ctx context.Context, credentialsFile, tokenFile string, scopes ...string) http.RoundTripper {
	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	if len(scopes) == 0 {
		scopes = []string{ScopeReadOnly}
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, tokenFile, config)
	return client.Transport
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, tokenFile string, config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(ctx, tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()

	_ = json.NewEncoder(f).Encode(token)
}
