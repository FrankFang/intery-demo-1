package gitee

import (
	"os"

	"golang.org/x/oauth2"
)

var giteeId = os.Getenv("GITEE_ID")
var giteeSecret = os.Getenv("GITEE_SECRET")

var Conf = &oauth2.Config{
	ClientID:     giteeId,
	ClientSecret: giteeSecret,
	Scopes:       []string{"projects", "emails"},
	RedirectURL:  "http://interycode.com/gitee_callback",
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://gitee.com/oauth/authorize",
		TokenURL: "https://gitee.com/oauth/token",
	},
}
