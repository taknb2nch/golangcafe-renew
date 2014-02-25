// 事前に以下の事をしておくこと。
// go getコマンドでパッケージを取得しておくこと。
// go get code.google.com/p/goauth2/oauth
// go get code.google.com/p/google-api-go-client/calendar/v3
//
// Cloud ConsoleのCredentialsでClient IDを作成しておく。
// Cloud ConsoleのAPIsでAPIをONにしておくこと。
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/calendar/v3"
)

var (
	cachefile = "cache.json"

	scope = "https://www.googleapis.com/auth/calendar"
	// request_urlは使用するAPIのURLを指定して下さい。（この例ではCalendarList）
	request_url       = "https://www.googleapis.com/calendar/v3/users/me/calendarList"
	request_token_url = "https://accounts.google.com/o/oauth2/auth"
	auth_token_url    = "https://accounts.google.com/o/oauth2/token"

	// clientID、secret、はDevelopers ConsoleのCredentialsからコピー＆ペーストして下さい。
	clientId     = ""
	clientSecret = ""
	//
	redirectURL = "http://localhost:16061"
)

func main() {
	runtime.GOMAXPROCS(2)
	flag.Parse()

	var err error

	fmt.Println("Start Execute API")

	// 認証コードを引数で受け取る。
	code := flag.Arg(0)

	//
	checkClientIDandSecret()

	config := &oauth.Config{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scope:        scope,
		AuthURL:      request_token_url,
		TokenURL:     auth_token_url,
		TokenCache:   oauth.CacheFile(cachefile),
	}

	transport := &oauth.Transport{Config: config}

	// キャッシュからトークンファイルを取得
	_, err = config.TokenCache.Token()
	if err != nil {
		// キャッシュなし
		if code == "" {
			code, err = getAuthCode(config)

			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}

		// 認証トークンを取得する。（取得後、キャッシュへ）
		_, err = transport.Exchange(code)
		if err != nil {
			fmt.Printf("Exchange Error: %v\n", err)
			os.Exit(1)
		}
	}

	//
	var svc *calendar.Service
	var cl *calendar.CalendarList

	svc, err = calendar.New(transport.Client())

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	cl, err = svc.CalendarList.List().Do()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("--- Your calendars ---\n")

	for _, item := range cl.Items {
		fmt.Printf("%v, %v\n", item.Summary, item.Description)
	}
}

func getAuthCode(config *oauth.Config) (string, error) {
	url := config.AuthCodeURL("")

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		url = strings.Replace(url, "&", `^&`, -1)
		cmd = exec.Command("cmd", "/c", "start", url)

	case "darwin":
		url = strings.Replace(url, "&", `\&`, -1)
		cmd = exec.Command("open", url)

	default:
		return "", fmt.Errorf("ブラウザで以下のURLにアクセスし、認証して下さい。\n%s\n", url)
	}

	redirectResult := make(chan RedirectResult, 1)
	serverStarted := make(chan bool, 1)
	//
	go func(rr chan<- RedirectResult, ss chan<- bool) {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")

			if code == "" {
				rr <- RedirectResult{Err: fmt.Errorf("codeを取得できませんでした。")}
			}

			fmt.Fprintf(w, `<!doctype html>
<html lang="ja">
<head>
<meta charset="utf-8">
</head>
<body onload="window.open('about:blank','_self').close();">
ブラウザが自動で閉じない場合は手動で閉じてください。
</body>
</html>
`)
			rr <- RedirectResult{Code: code}
		})

		fmt.Printf("Start Listen: localhost:16061\n")
		ss <- true

		err := http.ListenAndServe(":16061", nil)

		if err != nil {
			rr <- RedirectResult{Err: err}
		}
	}(redirectResult, serverStarted)

	<-serverStarted

	// set redirect timeout 20sec
	tch := time.After(20 * time.Second)

	fmt.Println("Start your browser after 2sec.")

	time.Sleep(2 * time.Second)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("Browser Start Error: %v\n", err)
	}

	var rr RedirectResult

	select {
	case rr = <-redirectResult:
	case <-tch:
		return "", fmt.Errorf("Timeout: waiting redirect.")
	}

	if rr.Err != nil {
		return "", fmt.Errorf("Redirect Error: %v\n", rr.Err)
	}

	fmt.Printf("Got code.\n")

	return rr.Code, nil
}

type RedirectResult struct {
	Code string
	Err  error
}

func checkClientIDandSecret() {
	if _, err := os.Stat(cachefile); err == nil {
		return
	}

	if clientId == "" {
		fmt.Println("Input ClientID")
		fmt.Scanf("%s\n", &clientId)
	}

	if clientSecret == "" {
		fmt.Println("Input ClientSecret")
		fmt.Scanf("%s\n", &clientSecret)
	}
}
