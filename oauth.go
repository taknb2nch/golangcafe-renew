package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	CALLBACK_PARAM         = "oauth_callback"
	CONSUMER_KEY_PARAM     = "oauth_consumer_key"
	NONCE_PARAM            = "oauth_nonce"
	SESSION_HANDLE_PARAM   = "oauth_session_handle"
	SIGNATURE_METHOD_PARAM = "oauth_signature_method"
	SIGNATURE_PARAM        = "oauth_signature"
	TIMESTAMP_PARAM        = "oauth_timestamp"
	TOKEN_PARAM            = "oauth_token"
	TOKEN_SECRET_PARAM     = "oauth_token_secret"
	VERIFIER_PARAM         = "oauth_verifier"
	VERSION_PARAM          = "oauth_version"
)

func main() {
	var verifyUrl string
	var verifyCode string
	var err error

	config := &Config{
		ConsumerKey:     "g43py4E3BUGjXjOWhxWjg",
		ConsumerSecret:  "EdqZpFtdmDBnpdEtGcIOzNmUetj29FgVZ0jZQNxzk",
		RequestTokenUrl: "https://api.twitter.com/oauth/request_token",
		AuthorizeUrl:    "https://api.twitter.com/oauth/authorize",
		AccessTokenUrl:  "https://api.twitter.com/oauth/access_token",
		TokenCache:      CacheFile("cache.json"),
	}

	h := HogeHoge{Config: config}

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

type Config struct {
	ConsumerKey     string
	ConsumerSecret  string
	RequestTokenUrl string
	AuthorizeUrl    string
	AccessTokenUrl  string

	TokenCache CacheFile
}

type Token struct {
	OAuthToken       string
	OAuthTokenSecret string
}

type HogeHoge struct {
	Config *Config
	Token  *Token

	tempToken       string
	tempTokenSecret string
}

func (h *HogeHoge) GetAuthUrl() (string, error) {
	if err := h.step1(); err != nil {
		return "", err
	}

	return h.Config.AuthorizeUrl + "?oauth_token=" + h.tempToken, nil
}

func (h *HogeHoge) Exchange(verifyCode string) (*Token, error) {

	token, err := h.step3(verifyCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

/*
	・コンシューマキー ( oauth_consumer_key )
	・ユーザが承認したときに（最終的にコンシューマの使用を許可させるために）差し戻すURI ( oauth_callback )
	あと、必要なもの
	・OAuthのバージョン (oauth_version )
	・タイムスタンプ ( oauth_timestamp )
	・当該アクセスに対して、一意性を表す文字列( oauth_nonce )
	・署名のプロトコル( oauth_signature_method )
	・署名( oauth_signature )
*/
// 署名キー
// "Consumer SecretをURLエンコードした値"&"Token Secretの値"
/*
	・oauth_token ・・・ユーザのトークン（仮）
	・oauth_token_secret ・・・ ユーザの秘密鍵
	・oauth_callback_confirmed ・・・ OKだった場合これが「true」になる。なお、そもそも失敗した場合、HTTPプロトコルの「401」
*/
func (h *HogeHoge) step1() error {
	hh := NewHoge1(
		"GET",
		h.Config.RequestTokenUrl,
		h.Config.ConsumerKey,
		h.Config.ConsumerSecret,
		"",
		"")

	hh.Set(CALLBACK_PARAM, "oob")

	value, err := doRequest("GET", h.Config.RequestTokenUrl, hh.GetAuthorization(), nil, nil)
	if err != nil {
		return err
	}

	h.tempToken = value.Get("oauth_token")
	h.tempTokenSecret = value.Get("oauth_token_secret")

	return nil
}

// func step2(oauthToken string) string {
// 	return AUTHORIZE_URL + "?oauth_token=" + oauthToken
// }

/*
	・oauth_consumer_key ・・・ コンシューマのキー
	・oauth_token ・・・ステップ２で戻ってきた oauth_token（ユーザごとの、ね）
	あと、必要なもの
	・OAuthのバージョン (oauth_version )
	・タイムスタンプ ( oauth_timestamp )
	・当該アクセスに対して、一意性を表す文字列( oauth_nonce )
	・署名のプロトコル( oauth_signature_method )
	・署名( oauth_signature )
*/
/*
	・oauth_token ・・・ 『正式の』アクセストークン
	・oauth_token_secret ・・・『正式の』アクセストークン秘密鍵
*/
func (h *HogeHoge) step3(verifyCode string) (*Token, error) {
	hh := NewHoge1(
		"POST",
		h.Config.AccessTokenUrl,
		h.Config.ConsumerKey,
		h.Config.ConsumerSecret,
		h.tempToken,
		h.tempTokenSecret)

	hh.Set(CALLBACK_PARAM, "oob")
	hh.Set(VERIFIER_PARAM, verifyCode)

	value, err := doRequest("POST", h.Config.AccessTokenUrl, hh.GetAuthorization(), nil, nil)
	if err != nil {
		return nil, err
	}

	tok := &Token{
		OAuthToken:       value.Get("oauth_token"),
		OAuthTokenSecret: value.Get("oauth_token_secret"),
	}
	h.Config.TokenCache.PutToken(tok)

	h.tempToken = ""
	h.tempTokenSecret = ""

	return tok, nil
}

func (h *HogeHoge) DoRequest(method, requestUrl string, v url.Values) {
	hh := NewHoge1(
		method,
		requestUrl,
		h.Config.ConsumerKey,
		h.Config.ConsumerSecret,
		h.Token.OAuthToken,
		h.Token.OAuthTokenSecret)

	hh.SetValues(v)

	value, _ := doRequest(method, requestUrl, hh.GetAuthorization(), nil, v)

	fmt.Println(value)
}

func doRequest(method, requestUrl, oauthHeader string, header http.Header, values url.Values) (url.Values, error) {
	var err error
	var req *http.Request
	var res *http.Response

	client := &http.Client{}

	pd := ""
	if values != nil {
		s := ""
		for k, _ := range values {
			s += percentEncode(k) + "=" + percentEncode(values.Get(k)) + "&"
		}
		pd = s[:len(s)-1]
	}

	//fmt.Println(pd)

	if req, err = http.NewRequest(method, requestUrl, strings.NewReader(pd)); err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Authorization", oauthHeader)

	for k, _ := range header {
		req.Header.Add(k, header.Get(k))
	}

	if values != nil {
		req.Header.Add("Content-Length", strconv.FormatInt(int64(len(pd)), 10))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Add("Content-Length", "0")
	}

	//fmt.Println(req)

	res, err = client.Do(req)

	//fmt.Println(res)

	if !(res.StatusCode >= http.StatusOK && res.StatusCode <= http.StatusPartialContent) {
		return nil, errors.New(fmt.Sprintf("http Response Error [%d]", res.StatusCode))
	}

	// レスポンスを解析する。
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	value, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}

	//fmt.Println(value)
	return value, nil
}

func percentEncode(str string) string {
	s := url.QueryEscape(str)

	//.replace("+", "%20")
	//.replace("*", "%2A")
	//.replace("%7E", "~");
	s = strings.Replace(s, "+", "%20", -1)
	s = strings.Replace(s, "*", "%2A", -1)
	s = strings.Replace(s, "%7E", "~", -1)

	return s
}

// 以下のパラメータは自動的にセットされます。
// oauth_nonce
// oauth_signature_method
// oauth_timestamp
// oauth_version
type Hoge1 struct {
	Method string
	Url    string

	consumerKey      string
	consumerSecret   string
	oauthToken       string
	oauthTokenSecret string
	params           map[string]string
}

func NewHoge1(method, url, consumerKey, consumerSecret, oauthToken, oauthTokenSecret string) Hoge1 {
	h := Hoge1{
		Method:           method,
		Url:              url,
		consumerKey:      consumerKey,
		consumerSecret:   consumerSecret,
		oauthToken:       oauthToken,
		oauthTokenSecret: oauthTokenSecret,
	}
	h.Clear()

	return h
}

func (h *Hoge1) Clear() {
	h.params = make(map[string]string)
}

func (h *Hoge1) Set(k, v string) {
	h.params[percentEncode(k)] = percentEncode(v)
}

func (h *Hoge1) SetValues(values url.Values) {
	for k, _ := range values {
		h.Set(k, values.Get(k))
	}
}

func (h *Hoge1) GetAuthorization() string {
	h.setDefaultParams()

	h.Set("oauth_signature", h.calcSignature())

	mk := h.sortedKeys(h.params)

	s := "OAuth "

	for _, k := range mk {
		s += k + "=\"" + h.params[k] + "\", "
	}

	return s[:len(s)-1]
}

func (h *Hoge1) setDefaultParams() {
	now := time.Now()
	timestamp := now.Unix()
	nonce := rand.New(rand.NewSource(now.UnixNano()))

	h.Set(CONSUMER_KEY_PARAM, h.consumerKey)

	if h.oauthToken != "" {
		h.Set(TOKEN_PARAM, h.oauthToken)
	}

	h.Set(NONCE_PARAM, strconv.FormatInt(nonce.Int63(), 10))
	h.Set(SIGNATURE_METHOD_PARAM, "HMAC-SHA1")
	h.Set(TIMESTAMP_PARAM, strconv.FormatInt(timestamp, 10))
	h.Set(VERSION_PARAM, "1.0")
}

func (h *Hoge1) calcSignature() string {
	mk := h.sortedKeys(h.params)

	// Parameter string
	ps := ""

	for _, k := range mk {
		ps += k + "=" + h.params[k] + "&"
	}

	// Signature base string
	sbs := percentEncode(h.Method) + "&" + percentEncode(h.Url) + "&" + percentEncode(ps[:len(ps)-1])

	//fmt.Println(sbs)

	// Signing key
	signingKey := percentEncode(h.consumerSecret) + "&" + percentEncode(h.oauthTokenSecret)

	//fmt.Println(signingKey)

	hashfun := hmac.New(sha1.New, []byte(signingKey))
	hashfun.Write([]byte(sbs))

	rawsignature := hashfun.Sum(nil)

	// base64エンコード
	base64signature := make([]byte, base64.StdEncoding.EncodedLen(len(rawsignature)))
	base64.StdEncoding.Encode(base64signature, rawsignature)

	//fmt.Println(string(base64signature))

	return string(base64signature)
}

func (h *Hoge1) sortedKeys(m map[string]string) []string {
	mk := make([]string, len(m))
	i := 0
	for k, _ := range m {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	return mk
}

type OAuthError struct {
	prefix string
	msg    string
}

func (oe OAuthError) Error() string {
	return "OAuthError: " + oe.prefix + ": " + oe.msg
}

type Cache interface {
	Token() (*Token, error)
	PutToken(*Token) error
}

type CacheFile string

func (f CacheFile) Token() (*Token, error) {
	file, err := os.Open(string(f))
	if err != nil {
		return nil, OAuthError{"CacheFile.Token", err.Error()}
	}
	defer file.Close()
	tok := &Token{}
	if err := json.NewDecoder(file).Decode(tok); err != nil {
		return nil, OAuthError{"CacheFile.Token", err.Error()}
	}
	return tok, nil
}

func (f CacheFile) PutToken(tok *Token) error {
	file, err := os.OpenFile(string(f), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return OAuthError{"CacheFile.PutToken", err.Error()}
	}
	if err := json.NewEncoder(file).Encode(tok); err != nil {
		file.Close()
		return OAuthError{"CacheFile.PutToken", err.Error()}
	}
	if err := file.Close(); err != nil {
		return OAuthError{"CacheFile.PutToken", err.Error()}
	}
	return nil
}
