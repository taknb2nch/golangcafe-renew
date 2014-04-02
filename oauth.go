package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	var verifyCode string
	var oauthToken string
	var oauthTokenSecret string

	v := step1()
	oauthToken = v.Get("oauth_token")
	oauthTokenSecret = v.Get("oauth_token_secret")

	url := step2(oauthToken)

	fmt.Println(url)
	fmt.Scanf("%s\n", &verifyCode)

	v = step3(verifyCode, oauthToken, oauthTokenSecret)
	oauthToken = v.Get("oauth_token")
	oauthTokenSecret = v.Get("oauth_token_secret")

	step4(oauthToken, oauthTokenSecret)
}

const (
	consumerKey    = "g43py4E3BUGjXjOWhxWjg"
	consumerSecret = "EdqZpFtdmDBnpdEtGcIOzNmUetj29FgVZ0jZQNxzk"

	OAUTH_VERSION    = "1.0"
	SIGNATURE_METHOD = "HMAC-SHA1"

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

	REQUEST_TOKEN_URL = "https://api.twitter.com/oauth/request_token"
	AUTHORIZE_URL     = "https://api.twitter.com/oauth/authorize"
	ACCESS_TOKEN_URL  = "https://api.twitter.com/oauth/access_token"
)

func step1() url.Values {
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

	now := time.Now()
	timestamp := now.Unix()
	nonce := rand.New(rand.NewSource(now.UnixNano()))

	h := NewHoge1(
		"GET",
		REQUEST_TOKEN_URL,
		consumerSecret,
		"")

	h.Set(NONCE_PARAM, strconv.FormatInt(nonce.Int63(), 10))
	h.Set(VERSION_PARAM, OAUTH_VERSION)
	h.Set(TIMESTAMP_PARAM, strconv.FormatInt(timestamp, 10))
	h.Set(SIGNATURE_METHOD_PARAM, SIGNATURE_METHOD)

	h.Set(CONSUMER_KEY_PARAM, consumerKey)

	h.Set(CALLBACK_PARAM, "oob")

	oauthHeader := h.Get()

	value, _ := doRequest("GET", REQUEST_TOKEN_URL, oauthHeader, nil, nil)

	/*
		・oauth_token ・・・ユーザのトークン（仮）
		・oauth_token_secret ・・・ ユーザの秘密鍵
		・oauth_callback_confirmed ・・・ OKだった場合これが「true」になる。なお、そもそも失敗した場合、HTTPプロトコルの「401」
	*/
	return value
}

func step2(oauthToken string) string {
	return AUTHORIZE_URL + "?oauth_token=" + oauthToken
}

func step3(verifyCode, oauthToken, oauthTokenSecret string) url.Values {
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
	now := time.Now()
	timestamp := now.Unix()
	nonce := rand.New(rand.NewSource(now.UnixNano()))

	h := NewHoge1(
		"POST",
		ACCESS_TOKEN_URL,
		consumerSecret,
		oauthTokenSecret)

	h.Set(NONCE_PARAM, strconv.FormatInt(nonce.Int63(), 10))
	h.Set(VERSION_PARAM, OAUTH_VERSION)
	h.Set(TIMESTAMP_PARAM, strconv.FormatInt(timestamp, 10))
	h.Set(SIGNATURE_METHOD_PARAM, SIGNATURE_METHOD)

	h.Set(CONSUMER_KEY_PARAM, consumerKey)
	h.Set(TOKEN_PARAM, oauthToken)

	h.Set(CALLBACK_PARAM, "oob")
	h.Set(VERIFIER_PARAM, verifyCode)

	oauthHeader := h.Get()

	value, _ := doRequest("POST", ACCESS_TOKEN_URL, oauthHeader, nil, nil)

	/*
		・oauth_token ・・・ 『正式の』アクセストークン
		・oauth_token_secret ・・・『正式の』アクセストークン秘密鍵
	*/
	return value
}

func step4(oauthToken, oauthTokenSecret string) {
	v := make(url.Values)
	//v.Add("status", "Hello Ladies + Gentlemen, a signed OAuth request!")
	v.Add("status", "post from go program. "+time.Now().String()+" #gdgchugoku")
	v.Add("lat", "37.7821120598956")
	v.Add("long", "-122.400612831116")

	now := time.Now()
	timestamp := now.Unix()
	nonce := rand.New(rand.NewSource(now.UnixNano()))

	h := NewHoge1(
		"POST",
		"https://api.twitter.com/1.1/statuses/update.json",
		consumerSecret,
		oauthTokenSecret)

	h.Set(NONCE_PARAM, strconv.FormatInt(nonce.Int63(), 10))
	h.Set(VERSION_PARAM, OAUTH_VERSION)
	h.Set(TIMESTAMP_PARAM, strconv.FormatInt(timestamp, 10))
	h.Set(SIGNATURE_METHOD_PARAM, SIGNATURE_METHOD)

	h.Set(CONSUMER_KEY_PARAM, consumerKey)
	h.Set(TOKEN_PARAM, oauthToken)

	h.SetValues(v)

	oauthHeader := h.Get()

	doRequest("POST", "https://api.twitter.com/1.1/statuses/update.json", oauthHeader, nil, v)
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

	fmt.Println(pd)

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

	fmt.Println(value)
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

type Hoge1 struct {
	Method           string
	Url              string
	ConsumerSecret   string
	OAuthTokenSecret string
	params           map[string]string
}

func NewHoge1(method, url, consumerSecret, tokenSecret string) Hoge1 {
	h := Hoge1{
		Method:           method,
		Url:              url,
		ConsumerSecret:   consumerSecret,
		OAuthTokenSecret: tokenSecret,
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

func (h *Hoge1) Get() string {
	h.params[percentEncode("oauth_signature")] = percentEncode(h.calcSignature())

	mk := h.sortedKeys(h.params)

	s := "OAuth "

	for _, k := range mk {
		s += k + "=\"" + h.params[k] + "\", "
	}

	return s[:len(s)-1]
}

func (h *Hoge1) calcSignature() string {
	mk := h.sortedKeys(h.params)

	ss := ""

	for _, k := range mk {
		ss += k + "=" + h.params[k] + "&"
	}

	s := percentEncode(h.Method) + "&" + percentEncode(h.Url) + "&" + percentEncode(ss[:len(ss)-1])

	//fmt.Println(s)

	signingKey := percentEncode(h.ConsumerSecret) + "&" + percentEncode(h.OAuthTokenSecret)

	//fmt.Println(signingKey)

	hashfun := hmac.New(sha1.New, []byte(signingKey))
	hashfun.Write([]byte(s))

	rawsignature := hashfun.Sum(nil)

	// base64エンコード
	base64signature := make([]byte, base64.StdEncoding.EncodedLen(len(rawsignature)))
	base64.StdEncoding.Encode(base64signature, rawsignature)

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
