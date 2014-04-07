package oauth

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
	callbackParam        = "oauth_callback"
	consumerKeyParam     = "oauth_consumer_key"
	nonceParam           = "oauth_nonce"
	signatureParam       = "oauth_signature"
	signatureMethodParam = "oauth_signature_method"
	timestampParam       = "oauth_timestamp"
	tokenParam           = "oauth_token"
	tokenSecretParam     = "oauth_token_secret"
	verifierParam        = "oauth_verifier"
	versionParam         = "oauth_version"
	oauthTokenKey        = "oauth_token"
	oauthTokenSecretKey  = "oauth_token_secret"
)

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

type Transport struct {
	*Config
	*Token

	tempToken       string
	tempTokenSecret string
}

func (h *Transport) GetAuthUrl() (string, error) {
	if err := h.doRequestToken(); err != nil {
		return "", err
	}

	return h.Config.AuthorizeUrl + "?" + tokenParam + "=" + h.tempToken, nil
}

func (h *Transport) Exchange(verifyCode string) (*Token, error) {
	token, err := h.doAccessToken(verifyCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

//
func (h *Transport) doRequestToken() error {
	hh := NewGenerator(
		h.ConsumerKey,
		h.ConsumerSecret,
		"",
		"")

	hh.SetUrl("GET", h.RequestTokenUrl, nil)

	hh.SetParam(callbackParam, "oob")

	value, err := doRequest("GET", h.RequestTokenUrl, hh.GetAuthorization(), nil, nil)
	if err != nil {
		return err
	}

	h.tempToken = value.Get(oauthTokenKey)
	h.tempTokenSecret = value.Get(oauthTokenSecretKey)

	return nil
}

//
func (h *Transport) doAccessToken(verifyCode string) (*Token, error) {
	hh := NewGenerator(
		h.ConsumerKey,
		h.ConsumerSecret,
		h.tempToken,
		h.tempTokenSecret)

	hh.SetUrl("POST", h.AccessTokenUrl, nil)

	hh.SetParam(callbackParam, "oob")
	hh.SetParam(verifierParam, verifyCode)

	value, err := doRequest("POST", h.AccessTokenUrl, hh.GetAuthorization(), nil, nil)
	if err != nil {
		return nil, err
	}

	tok := &Token{
		OAuthToken:       value.Get(oauthTokenKey),
		OAuthTokenSecret: value.Get(oauthTokenSecretKey),
	}
	h.TokenCache.PutToken(tok)

	h.tempToken = ""
	h.tempTokenSecret = ""

	return tok, nil
}

func (h *Transport) DoRequest(method, requestUrl string, v url.Values) {
	hh := NewGenerator(
		h.ConsumerKey,
		h.ConsumerSecret,
		h.Token.OAuthToken,
		h.Token.OAuthTokenSecret)

	hh.SetUrl(method, requestUrl, v)

	value, _ := doRequest(method, requestUrl, hh.GetAuthorization(), nil, v)

	fmt.Println(value)
}

func (h *Transport) GetGenerator() Generator {
	hh := NewGenerator(
		h.ConsumerKey,
		h.ConsumerSecret,
		h.OAuthToken,
		h.OAuthTokenSecret)

	return hh
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
			s += PercentEncode(k) + "=" + PercentEncode(values.Get(k)) + "&"
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

// Authorization header generator.
// Following parameters are set automatically, when calculating.
// oauth_nonce
// oauth_signature_method
// oauth_timestamp
// oauth_version
type Generator struct {
	method           string
	requestUrl       string
	consumerKey      string
	consumerSecret   string
	oauthToken       string
	oauthTokenSecret string
	params           map[string]string

	calcTimeAndNonce func() (string, string)
}

// NewGenerator returns a new Generator.
// Specify "", if oauthTokenSecret and oauthToken are not exist.
func NewGenerator(consumerKey, consumerSecret, oauthToken, oauthTokenSecret string) Generator {
	h := Generator{
		consumerKey:      consumerKey,
		consumerSecret:   consumerSecret,
		oauthToken:       oauthToken,
		oauthTokenSecret: oauthTokenSecret,
	}
	h.clear()

	return h
}

func (h *Generator) clear() {
	h.params = make(map[string]string)

	h.calcTimeAndNonce = func() (string, string) {
		now := time.Now()

		return strconv.FormatInt(now.Unix(), 10), strconv.FormatInt(rand.New(rand.NewSource(now.UnixNano())).Int63(), 10)
	}
}

// Set sets the key to value. It replaces any existing values.
func (h *Generator) SetParam(k, v string) {
	h.params[PercentEncode(k)] = PercentEncode(v)
}

// SetUrl sets the method, requestUrl and values.
func (h *Generator) SetUrl(method, requestUrl string, values url.Values) error {
	h.method = method

	u, err := url.Parse(requestUrl)
	if err != nil {
		return err
	}

	h.requestUrl = u.Scheme + "://" + u.Host + u.Path
	qs := u.Query()
	for k, _ := range qs {
		h.SetParam(k, qs.Get(k))
	}

	if values != nil {
		for k, _ := range values {
			h.SetParam(k, values.Get(k))
		}
	}

	return nil
}

// GetAuthorization returns the value for Authorization header.
func (h *Generator) GetAuthorization() string {
	h.setDefaultParams()

	h.SetParam(signatureParam, h.calcSignature())

	mk := h.sortedKeys(h.params)

	s := "OAuth "

	for _, k := range mk {
		s += k + "=\"" + h.params[k] + "\", "
	}

	return s[:len(s)-1]
}

func (h *Generator) setDefaultParams() {
	timestamp, nonce := h.calcTimeAndNonce()

	h.SetParam(consumerKeyParam, h.consumerKey)

	if h.oauthToken != "" {
		h.SetParam(tokenParam, h.oauthToken)
	}

	h.SetParam(nonceParam, nonce)
	h.SetParam(signatureMethodParam, "HMAC-SHA1")
	h.SetParam(timestampParam, timestamp)
	h.SetParam(versionParam, "1.0")
}

func (h *Generator) calcSignature() string {
	mk := h.sortedKeys(h.params)

	// Parameter string
	ps := ""

	for _, k := range mk {
		ps += k + "=" + h.params[k] + "&"
	}

	//fmt.Println(ps)

	// Signature base string
	sbs := PercentEncode(h.method) + "&" + PercentEncode(h.requestUrl) + "&" + PercentEncode(ps[:len(ps)-1])

	//fmt.Println(sbs)

	// Signing key
	signingKey := PercentEncode(h.consumerSecret) + "&" + PercentEncode(h.oauthTokenSecret)

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

func (h *Generator) sortedKeys(m map[string]string) []string {
	mk := make([]string, len(m))
	i := 0
	for k, _ := range m {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	return mk
}

// PercentEncode encodes the string so it can be safely placed inside a URL query.
func PercentEncode(str string) string {
	s := url.QueryEscape(str)
	s = strings.Replace(s, "+", "%20", -1)
	s = strings.Replace(s, "*", "%2A", -1)
	s = strings.Replace(s, "%7E", "~", -1)

	return s
}

type Cache interface {
	Token() (*Token, error)
	PutToken(*Token) error
}

type CacheFile string

func (f CacheFile) Token() (*Token, error) {
	var file *os.File
	var err error

	file, err = os.Open(string(f))

	if err != nil {
		return nil, err
	}

	defer file.Close()

	tok := &Token{}

	if err = json.NewDecoder(file).Decode(tok); err != nil {
		return nil, err
	}

	return tok, nil
}

func (f CacheFile) PutToken(tok *Token) error {
	var file *os.File
	var err error

	file, err = os.OpenFile(string(f), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return err
	}

	if err = json.NewEncoder(file).Encode(tok); err != nil {
		file.Close()
		return err
	}

	if err = file.Close(); err != nil {
		return err
	}

	return nil
}
