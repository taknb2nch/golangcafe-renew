package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/taknb2nch/goauth"
)

func main() {
	var verifyUrl string
	var verifyCode string
	var err error

	config := &oauth.Config{
		ConsumerKey:     "g43py4E3BUGjXjOWhxWjg",
		ConsumerSecret:  "EdqZpFtdmDBnpdEtGcIOzNmUetj29FgVZ0jZQNxzk",
		RequestTokenUrl: "https://api.twitter.com/oauth/request_token",
		AuthorizeUrl:    "https://api.twitter.com/oauth/authorize",
		AccessTokenUrl:  "https://api.twitter.com/oauth/access_token",
		TokenCache:      oauth.CacheFile("cache.json"),
	}

	h := oauth.Transport{Config: config}

	token, err := config.TokenCache.Token()

	if err != nil {
		verifyUrl, err = h.GetAuthUrl()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Visit this URL to get a code, then enter below this.\n")
		fmt.Println(verifyUrl)
		fmt.Printf("> ")
		fmt.Scanf("%s\n", &verifyCode)

		token, err = h.Exchange(verifyCode)
		if err != nil {
			log.Fatalln(err)
		}
	}

	h.Token = token

	v := make(url.Values)
	//v.Add("status", "Hello Ladies + Gentlemen, a signed OAuth request!")
	v.Add("status", "post from go program. "+time.Now().String()+" #gdgchugoku")
	v.Add("lat", "37.7821120598956")
	v.Add("long", "-122.400612831116")

	h.DoRequest("POST", "https://api.twitter.com/1.1/statuses/update.json", v)
}
