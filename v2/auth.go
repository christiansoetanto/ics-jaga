package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const tokenFile = "token.json"

func newCalendarService(ctx context.Context) *calendar.Service {
	logger.Infof("loading credentials from %q", config.CredsFile)
	b, err := os.ReadFile(config.CredsFile)
	if err != nil {
		logger.Fatalf("read credentials %q: %v\n  → download from Google Cloud Console → APIs & Services → Credentials → OAuth 2.0 Client (Desktop)", config.CredsFile, err)
	}

	oauthCfg, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		logger.Fatalf("parse credentials: %v", err)
	}

	client := oauthClient(oauthCfg)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logger.Fatalf("create calendar service: %v", err)
	}
	logger.Infof("calendar service ready")
	return srv
}

func oauthClient(cfg *oauth2.Config) *http.Client {
	tok, err := loadToken()
	if err != nil {
		logger.Infof("no cached token found, starting browser auth flow...")
		tok = authFromBrowser(cfg)
		saveToken(tok)
	} else {
		logger.Infof("loaded cached token from %q", tokenFile)
	}
	return cfg.Client(context.Background(), tok)
}

func loadToken() (*oauth2.Token, error) {
	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, err
	}
	logger.Debugf("token loaded: valid=%v expiry=%v", tok.Valid(), tok.Expiry)
	return tok, nil
}

func authFromBrowser(cfg *oauth2.Config) *oauth2.Token {
	authURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\nOpen this URL in your browser:\n\n  %s\n\nPaste the auth code: ", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		logger.Fatalf("read auth code: %v", err)
	}

	logger.Infof("exchanging auth code for token...")
	tok, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		logger.Fatalf("exchange auth code: %v", err)
	}
	logger.Infof("token obtained, expiry: %v", tok.Expiry)
	return tok
}

func saveToken(tok *oauth2.Token) {
	f, err := os.Create(tokenFile)
	if err != nil {
		logger.Fatalf("save token to %q: %v", tokenFile, err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(tok); err != nil {
		logger.Fatalf("encode token: %v", err)
	}
	logger.Infof("token cached to %q", tokenFile)
}
