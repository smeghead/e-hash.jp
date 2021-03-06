package utamaru

import (
	"os"
	"errors"
	"math/big"
	"fmt"
	"bytes"
	"time"
	"crypto/rand"
	"strings"
	"strconv"
	"sort"
	"encoding/json"
	"net/http"
	"appengine"
	"appengine/urlfetch"
	"io/ioutil"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

var HashtagRexexp string = "[#＃][^ ;'.,　\n]+"

type User struct {
	Id_Str string
	Screen_name string
	Profile_Image_Url string
}

type TweetTw struct {
	Id_Str string
	User User
	From_User string
	From_User_Id_Str string
	To_User string
	To_User_Id_Str string
	Text string
	Profile_Image_Url string
	Created_At string
}

type ErrorResponse struct {
	Error string
}

func oAuthHeader(c appengine.Context, method, url string, options map[string]interface{}) string {
	withUserAccessToken := false
	var user *TwitterUser
	if options["withUserAccessToken"] != nil {
		withUserAccessToken = options["withUserAccessToken"].(bool)
		user = options["user"].(*TwitterUser)
	}
	withApplicationAccessToken := false
	if options["withApplicationAccessToken"] != nil {
		withApplicationAccessToken = options["withApplicationAccessToken"].(bool)
	}
	var requestToken *TwitterRequestToken
	if options["requestToken"] != nil {
		requestToken = options["requestToken"].(*TwitterRequestToken)
	}
	oauthVerifier := ""
	if options["oauthVerifier"] != nil {
		oauthVerifier = options["oauthVerifier"].(string)
	}
	status := ""
	if options["status"] != nil {
		status = options["status"].(string)
	}

	conf, err := GetTwitterConf(c)
	if err != nil {
		c.Errorf("oAuthHeader failed to load TwitterConf: %v", err)
		return ""
	}
	oauthMap := make(map[string]string)
	oauthMap["oauth_consumer_key"] = conf.ConsumerKey
	oauthMap["oauth_signature_method"] = "HMAC-SHA1"
	oauthMap["oauth_timestamp"] = strconv.Itoa(int(time.Now().Unix()))
	n, err := rand.Int(rand.Reader, big.NewInt(99999))
	oauthMap["oauth_nonce"] = strconv.Itoa(int(n.Int64()))
	oauthMap["oauth_version"] = "1.0"
	if withApplicationAccessToken {
		oauthMap["oauth_token"] = conf.AccessToken
	} else if withUserAccessToken {
		oauthMap["oauth_token"] = user.OauthToken
	} else if requestToken != nil {
		oauthMap["oauth_token"] = requestToken.OauthToken
		oauthMap["oauth_verifier"] = oauthVerifier
	}
	if len(status) > 0 {
		oauthMap["status"] = status
	}

	oauthArray := make([]string, 0, 10)
	mapKeys := make([]string, len(oauthMap))
	i := 0
	for k, _ := range oauthMap {
		mapKeys[i] = k
		i++
	}
	sort.Strings(mapKeys)
	for _, k := range mapKeys {
		oauthArray = append(oauthArray, k + Encode("=") + Encode(oauthMap[k]))
	}
	oauth := strings.Join(oauthArray, Encode("&"))
	msg := method + "&" + Encode(url) + "&" + oauth
	c.Debugf("msg: %s", msg)
	key := conf.ConsumerSecret + "&"
	if withApplicationAccessToken {
		key += conf.AccessTokenSecret
	} else if withUserAccessToken {
		key += user.OauthSecret
	} else if requestToken != nil {
		key += requestToken.OauthSecret
	}
	c.Debugf("key: %s", key)
	c.Debugf("msg: %s", msg)
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(msg))
	oauthMap["oauth_signature"] = base64.StdEncoding.EncodeToString(h.Sum(nil))
	c.Debugf("oauth_signature: %s", oauthMap["oauth_signature"])

	oauthArray = make([]string, 0, 10)
	for k, v := range oauthMap {
		if k == "status" {
			continue
		}
		oauthArray = append(oauthArray, k + "=\"" + Encode(v) + "\"")
	}
	oauth = strings.Join(oauthArray, ", ")
	return "OAuth " + oauth
}

func GetRequestToken(c appengine.Context) (*TwitterRequestToken, error) {
	url := "http://twitter.com/oauth/request_token"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", oAuthHeader(c, "GET", url, map[string]interface{}{}))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("GetRequestToken failed to api call: %v", err)
		return nil, err
	}
	bytes, err2 := ioutil.ReadAll(response.Body)
	c.Debugf("response body: %s", string(bytes))
	c.Debugf("response StatusCode: %d", response.StatusCode)
	if response.StatusCode != http.StatusOK {
		c.Errorf("GetRequestToken failed to get token. bad status code: %d", response.StatusCode)
		return nil, errors.New("GetRequestToken failed to get token.")
	}
	if err2 != nil {
		c.Errorf("GetRequestToken failed to read result: %v", err)
		return nil, err
	}
	s := string(bytes)
	splited := strings.Split(s, "&")
	resMap := make(map[string]string, 3)
	for _, e := range splited {
		c.Debugf("elem: %s", e)
		keyValue := strings.Split(e, "=")
		if len(keyValue) == 2 {
			c.Debugf("keyvalue: %s", keyValue)
			resMap[keyValue[0]] = keyValue[1]
		}
	}
	c.Debugf("resMap: %v", resMap)
	requestToken := NewRequestToken(resMap)
	return &requestToken, nil
}

func GetAccessToken(c appengine.Context, requestToken TwitterRequestToken, oauthVerifier string) (*TwitterUser, error) {
	url := "http://twitter.com/oauth/access_token"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", oAuthHeader(c, "GET", url, map[string]interface{}{
		"requestToken": &requestToken,
		"oauthVerifier": oauthVerifier,
	}))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("GetPublicTimeline failed to api call: %v", err)
		return nil, err
	}
	bytes, err2 := ioutil.ReadAll(response.Body)
	c.Debugf("response body: %s", string(bytes))
	if err2 != nil {
		c.Errorf("GetPublicTimeline failed to read result: %v", err)
		return nil, err
	}
	s := string(bytes)
	splited := strings.Split(s, "&")
	resMap := make(map[string]string, 3)
	for _, e := range splited {
		c.Debugf("elem: %s", e)
		keyValue := strings.Split(e, "=")
		if len(keyValue) == 2 {
			c.Debugf("keyvalue: %s", keyValue)
			resMap[keyValue[0]] = keyValue[1]
		}
	}
	c.Debugf("resMap: %v", resMap)
	user := NewTwitterUser(resMap)
	return &user, nil
}

func GetPublicTimeline(c appengine.Context) ([]TweetTw, error) {
	url := "http://api.twitter.com/statuses/public_timeline.json"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", oAuthHeader(c, "GET", url, map[string]interface{}{
		"withApplicationAccessToken": true,
	}))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("GetPublicTimeline failed to api call: %v", err)
		return nil, err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("GetPublicTimeline failed to read result: %v", err)
		return nil, err
	}
	tweets := make([]TweetTw, 0, 20)
	json.Unmarshal(jsonVal, &tweets)
	if len(tweets) == 0 {
		//maybe error.
		var errRes ErrorResponse
		json.Unmarshal(jsonVal, &errRes)
		c.Errorf("GetPublicTimeline error message: %v", errRes.Error)
		return nil, errors.New(errRes.Error)
	}
	for _, tweet := range tweets {
		c.Debugf("GetPublicTimeline %v: %v", tweet.User.Screen_name, tweet.Text)
	}
	return tweets, nil
}

func InvokePublicTimelineStream(c appengine.Context, procTweet func(TweetTw) error) error {
	request, _ := http.NewRequest(
			 "POST",
			 "http://stream.twitter.com/1/statuses/filter.json",
			 bytes.NewBufferString("track=#,＃"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Basic aW1ha2FyYXlhcnU6N2tvcm9iaThva2k=")
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("InvokePublicTimelineStream failed to call api: %v", err)
		return err
	}
	const NBUF = 512
	var buf [NBUF]byte
	var str string
	for {
		switch nr, er := response.Body.Read(buf[:]); true {
		case nr < 0:
			c.Errorf("InvokePublicTimelineStream read to api result: %v", er)
			os.Exit(1)
		case nr == 0: // EOF
			return nil
		case nr > 0:
			str += string(buf[0:nr])
			tweetsStrings := strings.Split(str, "\n")
			str = tweetsStrings[len(tweetsStrings) - 1]
			if len(tweetsStrings) > 1 {
				for _, e := range tweetsStrings {
					var tweet TweetTw
					json.Unmarshal(bytes.NewBufferString(e).Bytes(), &tweet)
					go procTweet(tweet)
				}
			}
		}
	}
	return nil
}

type Locale struct {
	Trends []Trend
}
type Trend struct {
	Name string
}

func GetTrends(c appengine.Context) ([]Trend, error) {
	url := "http://api.twitter.com/1/trends/1118370.json"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", oAuthHeader(c, "GET", url, map[string]interface{}{
		"withApplicationAccessToken": true,
	}))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("GetTrends failed to api call: %v", err)
		return nil, err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("GetTrends failed to read result: %v", err)
		return nil, err
	}
	locales := make([]Locale, 0, 1)
	json.Unmarshal(jsonVal, &locales)
	if len(locales) == 0 || len(locales[0].Trends) == 0 {
		//maybe error.
		var errRes ErrorResponse
		json.Unmarshal(jsonVal, &errRes)
		return nil, errors.New(errRes.Error)
	}
	for _, trend := range locales[0].Trends {
		c.Debugf("GetTrends trend.Name: %v", trend.Name)
	}
	return locales[0].Trends, nil
}

func HomeTest(c appengine.Context) error {
	url := "http://api.twitter.com/statuses/friends_timeline.json"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", oAuthHeader(c, "GET", url, map[string]interface{}{
		"withApplicationAccessToken": true,
	}))
	c.Debugf("HomeTest Authorization: %v", request.Header.Get("Authorization"))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("HomeTest failed to api call: %v", err)
		return err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("HomeTest failed to read result: %v", err)
		return err
	}
	c.Debugf("HomeTest response: %v", string(jsonVal))
	return nil
}

func Encode(s string) string {
	var enc string
	for _, c := range []byte(s) {
		if isEncodable(c) {
			enc += "%"
			enc += string("0123456789ABCDEF"[c>>4])
			enc += string("0123456789ABCDEF"[c&15])
		} else {
			enc += string(c)
		}
	}
	return enc
}

func isEncodable(c byte) bool {
	// return false if c is an unreserved character (see RFC 3986 section 2.3)
	switch {
	case (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
		return false
	case c >= '0' && c <= '9':
		return false
	case c == '-' || c == '.' || c == '_' || c == '~':
		return false
	}
	return true
}

type Result struct {
	Results []TweetTw
}

func SearchTweetsByHashtag(c appengine.Context, hashtag string) ([]TweetTw, error) {
	var empty []TweetTw
	if len(hashtag) == 0 {
		return empty, nil
	}
	h, _ := FindHashtag(c, hashtag)
	url := "http://e-hash.jp/p.php?rpp=100&q=" + Encode(hashtag)
	//url := "http://search.twitter.com/search.json?rpp=100&q=" + Encode(hashtag)
	if h.LastStatusId != "" {
		url += "&since_id=" + h.LastStatusId
	}
	c.Debugf("SearchTweetsByHashtag url: %s", url)
	request, _ := http.NewRequest("GET", url, nil)
//	request.Header.Set("Authorization", oAuthHeader(c, "GET", url, true, nil, ""))
//	c.Debugf("SearchTweetsByHashtag Authorization: %v", request.Header.Get("Authorization"))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("SearchTweetsByHashtag failed to api call: %v", err)
		return nil, err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("SearchTweetsByHashtag failed to read result: %v", err)
		return nil, err
	}
	var result Result
	json.Unmarshal(jsonVal, &result)
	if len(result.Results) == 0 {
		//maybe error.
		var errRes ErrorResponse
		json.Unmarshal(jsonVal, &errRes)
		if len(errRes.Error) > 0 {
			return nil, errors.New(errRes.Error)
		}
		// simply no results. not error
		return result.Results, nil
	}
//	for _, tweet := range result.Results {
//		c.Debugf("SearchTweetsByHashtag tweet.Text: %v", tweet.Text)
//	}
	return result.Results, nil
}

func PostTweet(c appengine.Context, status string) error {
	encodedStatus := Encode(status)
	url := "http://api.twitter.com/statuses/update.json"
	body := strings.NewReader("status=" + encodedStatus)
	request, _ := http.NewRequest("POST", url, body)
	request.Header.Set("Authorization", oAuthHeader(c, "POST", url, map[string]interface{}{
		"withApplicationAccessToken": true,
		"status": encodedStatus,
	}))
	c.Debugf("PostTweet Authorization: %v", request.Header.Get("Authorization"))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("PostTweet failed to api call: %v", err)
		return err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("PostTweet failed to read result: %v", err)
		return err
	}
	c.Debugf("PostTweet response: %v", string(jsonVal))
	if response.StatusCode != 200 {
		c.Errorf("PostTweet failed to post status. StatusCode: %d", response.StatusCode)
		return errors.New("PostTweet failed to post status.")
	}
	return nil
}

func PostTweetByUser(c appengine.Context, status string, user TwitterUser) (*TweetTw, error) {
	encodedStatus := Encode(status)
	url := "http://api.twitter.com/statuses/update.json"
	body := strings.NewReader("status=" + encodedStatus)
	request, _ := http.NewRequest("POST", url, body)
	request.Header.Set("Authorization", oAuthHeader(c, "POST", url, map[string]interface{}{
		"withUserAccessToken": true,
		"user": &user,
		"status": encodedStatus,
	}))
	c.Debugf("PostTweetByUser Authorization: %v", request.Header.Get("Authorization"))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("PostTweetByUser failed to api call: %v", err)
		return nil, err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("PostTweetByUser failed to read result: %v", err)
		return nil, err
	}
	c.Debugf("PostTweetByUser response: %v", string(jsonVal))
	if response.StatusCode != 200 {
		c.Errorf("PostTweetByUser failed to post status. StatusCode: %d", response.StatusCode)
		return nil, errors.New("PostTweetByUser failed to post status.")
	}
	var tweet TweetTw
	json.Unmarshal(jsonVal, &tweet)
	return &tweet, nil
}

func FavoriteTweetByUser(c appengine.Context, statusId string, user TwitterUser) error {
	url := fmt.Sprintf("http://api.twitter.com/1/favorites/create/%s.json", statusId)
	request, _ := http.NewRequest("POST", url, bytes.NewBufferString(""))
	request.Header.Set("Authorization", oAuthHeader(c, "POST", url, map[string]interface{}{
		"withUserAccessToken": true,
		"user": &user,
	}))
	c.Debugf("FavoriteTweetByUser Authorization: %v", request.Header.Get("Authorization"))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("FavoriteTweetByUser failed to api call: %v", err)
		return err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("FavoriteTweetByUser failed to read result: %v", err)
		return err
	}
	c.Debugf("FavoriteTweetByUser response: %v", string(jsonVal))
	if response.StatusCode != 200 {
		c.Errorf("FavoriteTweetByUser failed to post status. StatusCode: %d", response.StatusCode)
		return errors.New("FavoriteTweetByUser failed to post status.")
	}
	return nil
}

func RetweetTweetByUser(c appengine.Context, statusId string, user TwitterUser) error {
	url := fmt.Sprintf("http://api.twitter.com/1/statuses/retweet/%s.json", statusId)
	c.Debugf("RetweetTweetByUser url: %s", url)
	request, _ := http.NewRequest("POST", url, bytes.NewBufferString(""))
	request.Header.Set("Authorization", oAuthHeader(c, "POST", url, map[string]interface{}{
		"withUserAccessToken": true,
		"user": &user,
	}))
	c.Debugf("RetweetTweetByUser Authorization: %v", request.Header.Get("Authorization"))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("RetweetTweetByUser failed to api call: %v", err)
		return err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("RetweetTweetByUser failed to read result: %v", err)
		return err
	}
	c.Debugf("RetweetTweetByUser response: %v", string(jsonVal))
	if response.StatusCode != 200 {
		c.Errorf("RetweetTweetByUser failed to post status. StatusCode: %d", response.StatusCode)
		return errors.New("RetweetTweetByUser failed to post status.")
	}
	return nil
}

