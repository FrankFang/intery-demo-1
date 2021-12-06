package github

import (
	"fmt"
	"intery/server/models"
	"io/ioutil"
	"log"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Controller struct{}

var conf = &oauth2.Config{
	ClientID:     "c509b5c3f08700791d87",
	ClientSecret: "cdcd8662ff64410639c068c2eab51e2879060ecb",
	Scopes:       []string{"repo"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
}

func (ctrl Controller) Show(c *gin.Context) {
	url := conf.AuthCodeURL(uniuri.New())
	c.JSON(200, gin.H{
		"url": url,
	})
}

func (ctrl Controller) Callback(c *gin.Context) {
	code, hasCode := c.GetQuery("code")
	if !hasCode {
		c.JSON(400, gin.H{
			"reason": "no code",
		})
		return
	}
	tok, err := conf.Exchange(c, code)
	if err != nil {
		fmt.Println(err)
	}
	go func() {
		a := models.Authorization{}
		err := a.Save()
		if err != nil {
			log.Fatal(err)
		}
	}()

	client := conf.Client(c, tok)
	response, err := client.Get("https://api.github.com/user")
	if err != nil {
		fmt.Println(err)
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	w := c.Writer
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	_, _ = w.Write(bytes)
	/*
		{
			login: "FrankFang",
			id: 839559,
			node_id: "MDQ6VXNlcjgzOTU1OQ==",
			avatar_url: "https://avatars.githubusercontent.com/u/839559?v=4",
			gravatar_id: "",
			url: "https://api.github.com/users/FrankFang",
			html_url: "https://github.com/FrankFang",
			followers_url: "https://api.github.com/users/FrankFang/followers",
			following_url: "https://api.github.com/users/FrankFang/following{/other_user}",
			gists_url: "https://api.github.com/users/FrankFang/gists{/gist_id}",
			starred_url: "https://api.github.com/users/FrankFang/starred{/owner}{/repo}",
			subscriptions_url: "https://api.github.com/users/FrankFang/subscriptions",
			organizations_url: "https://api.github.com/users/FrankFang/orgs",
			repos_url: "https://api.github.com/users/FrankFang/repos",
			events_url: "https://api.github.com/users/FrankFang/events{/privacy}",
			received_events_url: "https://api.github.com/users/FrankFang/received_events",
			type: "User",
			site_admin: false,
			name: "Frank Fang",
			company: "@jirengu-inc ",
			blog: "https://fangyinghang.com/",
			location: "Hangzhou, China",
			email: null,
			hireable: null,
			bio: "Former Tencent Employee & Working at Alibaba.",
			twitter_username: null,
			public_repos: 366,
			public_gists: 94,
			followers: 2765,
			following: 112,
			created_at: "2011-06-09T09:16:40Z",
			updated_at: "2021-11-22T05:33:19Z"
			}
	*/
}
