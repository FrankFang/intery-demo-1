package github

import "golang.org/x/oauth2"

var Conf = &oauth2.Config{
	ClientID:     "c509b5c3f08700791d87",
	ClientSecret: "cdcd8662ff64410639c068c2eab51e2879060ecb",
	Scopes:       []string{"repo"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
}
