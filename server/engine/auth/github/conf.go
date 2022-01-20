package github

import (
	"os"

	"golang.org/x/oauth2"
)

var githubId = os.Getenv("GITHUB_ID")
var githubSecret = os.Getenv("GITHUB_SECRET")

var Conf = &oauth2.Config{
	ClientID:     githubId,
	ClientSecret: githubSecret,
	Scopes:       []string{"repo"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
}
