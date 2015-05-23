package google_auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/daemonl/go_gsd/shared"
)

type GoogleAuth struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	RedirectURI  string `json:"redirectURI"`
}

func (ga *GoogleAuth) TryToAuthenticate(request shared.IRequest) (string, interface{}, error) {
	googleToken, ok := request.QueryString().String("code")
	if !ok {
		return "", nil, nil
	}
	log.Printf("Google Auth with Token %s\n", googleToken)

	token, err := ga.Exchange(googleToken)
	if err != nil {
		return "", nil, err
	}

	email, err := token.GooglePlusID()
	if err != nil {
		return "", nil, err
	}

	log.Printf("Google auth response: %s\n", email)

	return "google_id", email, nil
}

// Token represents an OAuth token response.
type Token struct {
	AccessToken      string  `json:"access_token"`
	TokenType        string  `json:"token_type"`
	ExpiresIn        int     `json:"expires_in"`
	IdToken          string  `json:"id_token"`
	Error            *string `json:"error"`
	ErrorDescription string  `json:"error_description"`
}

type ClaimSet struct {
	Sub string
}

func base64Decode(s string) ([]byte, error) {
	// add back missing padding
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

func (t *Token) EmailAddress() (string, error) {
	return "NONE", nil
}

func (t *Token) GooglePlusID() (string, error) {
	var set ClaimSet
	if t.IdToken != "" {
		// Check that the padding is correct for a base64decode
		parts := strings.Split(t.IdToken, ".")
		if len(parts) < 2 {
			return "", fmt.Errorf("Malformed ID token")
		}
		// Decode the ID token
		fmt.Println(parts[1])
		b, err := base64Decode(parts[1])
		if err != nil {
			return "", fmt.Errorf("Malformed ID token: %v", err)
		}
		err = json.Unmarshal(b, &set)
		if err != nil {
			return "", fmt.Errorf("Malformed ID token: %v", err)
		}
	}
	return set.Sub, nil
}

func (ga *GoogleAuth) Revoke(code string) error {
	url := "https://accounts.google.com/o/oauth2/revoke?token=" + code
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to revoke token")
	}
	defer resp.Body.Close()
	return nil
}

func (ga *GoogleAuth) Exchange(code string) (*Token, error) {
	// Exchange the authorization code for a credentials object via a POST request
	addr := "https://accounts.google.com/o/oauth2/token"
	values := url.Values{
		"Content-Type":  {"application/x-www-form-urlencoded"},
		"code":          {code},
		"client_id":     {ga.ClientID},
		"client_secret": {ga.ClientSecret},
		"redirect_uri":  {ga.RedirectURI},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.PostForm(addr, values)
	if err != nil {
		return nil, fmt.Errorf("Exchanging code: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(bodyBytes))

	// Decode the response body into a token object
	var token Token
	err = json.Unmarshal(bodyBytes, &token)
	if err != nil {
		return nil, fmt.Errorf("Decoding access token: %v", err)
	}

	if token.Error != nil {
		return nil, fmt.Errorf(token.ErrorDescription)
	}

	return &token, nil
}
