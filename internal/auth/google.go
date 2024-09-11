package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var (
	home            = os.Getenv("HOME")
	credentialsPath = home + "/.sheets-agent"
	tokenFile       = credentialsPath + "/token.json"
	credentialsFile = credentialsPath + "/credentials.json"
	scopes          = []string{sheets.SpreadsheetsScope, sheets.DriveScope}
)

type GoogleAuth interface {
	GetClient() (string, error)
}

type serverAuthGoogle struct {
	server      *http.Server
	redirectURL string
	state       string
	err         error
	code        chan string
}

func NewServerAuthGoogle() *serverAuthGoogle {
	const redirectURL = "http://localhost:3000"
	return &serverAuthGoogle{
		server:      &http.Server{Addr: redirectURL},
		redirectURL: redirectURL,
		state:       randomStateString(32),
		code:        make(chan string, 1),
	}
}

// Reads the credentials from a file and returns an authenticated client.
func (s *serverAuthGoogle) GetClient() *http.Client {
	// Load the credentials file.
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Parse the credentials file.
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// You need a token file to store the token obtained from the OAuth process.
	tok, err := s.tokenFromFile(tokenFile)
	if err != nil {
		tok = s.getTokenFromWeb(config)
		s.saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Retrieves token from a local file.
func (s *serverAuthGoogle) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Obtains a token from the web using the OAuth2 process.
func (s *serverAuthGoogle) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	code, err := s.getCode(config)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Saves a token to a file path.
func (s *serverAuthGoogle) saveToken(path string, token *oauth2.Token) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (s *serverAuthGoogle) getCode(config *oauth2.Config) (string, error) {
	// parse the redirect URL for the port number
	u, err := url.Parse(s.redirectURL)
	if err != nil {
		return "", errors.Errorf("bad redirect URL: %s", err)
	}

	authorizationURL := config.AuthCodeURL(s.state, oauth2.AccessTypeOffline)
	// start a web server to listen on a callback URL
	http.HandleFunc("/google/auth/", s.handleAuthGoogle)

	// Write out helper prompts to the user including how to recover from the CLI "hanging" due to some issue
	// with browser opening, github login, github redir, http server setup
	fmt.Printf("\nYour browser has been opened to visit:\n%s\n\n", authorizationURL)
	fmt.Printf("Hit CTRL-C to exit CLI if something goes wrong ...\n\n")

	// set up a listener on the redirect port
	port := fmt.Sprintf(":%s", u.Port())
	l, err := net.Listen("tcp", port)
	if err != nil {
		return "", errors.Errorf("can't listen on port %s: %s", port, err)
	}

	// open a browser window to the authorizationURL
	if err := open.Start(authorizationURL); err != nil {
		return "", errors.Errorf("can't open browser to URL %s: %s", authorizationURL, err)
	}

	// start the blocking web server loop
	// this will exit when the handler gets fired and calls server.Close()
	_ = s.server.Serve(l) // ignore server closed error
	if s.err != nil {
		return "", s.err
	}
	// Waiting for a response from web server
	return <-s.code, nil
}

func (s *serverAuthGoogle) handleAuthGoogle(w http.ResponseWriter, r *http.Request) {
	// close the HTTP server
	defer s.cleanup(r.Context())
	// validate state
	retState := r.URL.Query().Get("state")
	if retState != s.state {
		s.err = fmt.Errorf("state fields do not match, this could be a security compromise attempt")
		writeMessage(w, "Error", "state fields do not match, this could be a security compromise attempt.")
		return
	}
	// get the authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		s.err = fmt.Errorf("url param 'code' is missing in %s", r.URL.RawQuery)
		writeMessage(w, "Error", "could not find 'code' URL parameter.")
		return
	}
	writeMessage(w, "Google access code generated successfully!", "You can close this window and return to the CLI.")
	s.code <- code
}

// cleanup closes the HTTP server
func (s *serverAuthGoogle) cleanup(ctx context.Context) {
	// we run this as a goroutine so that this function falls through and
	// the socket to the browser gets flushed/closed before the server goes away
	go func() {
		_ = s.server.Shutdown(ctx)
	}()
}
