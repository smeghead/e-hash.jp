package utamaru

import (
	"os"
	"bytes"
	"time"
	"rand"
	"strings"
	"strconv"
	"sort"
	"json"
	"http"
	"appengine"
	"appengine/urlfetch"
	"io/ioutil"
	"crypto/hmac"
	"encoding/base64"
)

var HashtagRexexp string = "[#＃][^ ;'.,　]+"

type User struct {
	Id_Str string
	Screen_name string
}

type TweetTw struct {
	Id_Str string
	User User
	From_User string
	Text string
	Profile_Image_Url string
	Created_At string
}

type ErrorResponse struct {
	Error string
}

func OAuthHeader(c appengine.Context, method, url string) string {
	conf, err := GetTwitterConf(c)
	if err != nil {
		c.Errorf("OAuthHeader failed to load TwitterConf: %v", err.String())
		return ""
	}
	oauthMap := make(map[string]string)
	oauthMap["oauth_consumer_key"] = conf.ConsumerKey
	oauthMap["oauth_signature_method"] = "HMAC-SHA1"
	oauthMap["oauth_timestamp"] = strconv.Itoa64(time.Seconds())
	oauthMap["oauth_nonce"] = strconv.Itoa64(rand.Int63())
	oauthMap["oauth_version"] = "1.0"
	oauthMap["oauth_token"] = conf.AccessToken

	oauthArray := make([]string, 0, 10)
	mapKeys := make([]string, len(oauthMap))
	i := 0
	for k, _ := range oauthMap {
		mapKeys[i] = k
		i++
	}
	sort.SortStrings(mapKeys)
	for _, k := range mapKeys {
		oauthArray = append(oauthArray, k + Encode("=") + Encode(oauthMap[k]))
	}
	oauth := strings.Join(oauthArray, Encode("&"))
	msg := method + "&" + Encode(url) + "&" + oauth
	key := conf.ConsumerSecret + "&" + conf.AccessTokenSecret
	h := hmac.NewSHA1([]byte(key))
	h.Write([]byte(msg))
	oauthMap["oauth_signature"] = base64.StdEncoding.EncodeToString(h.Sum())

	oauthArray = make([]string, 0, 10)
	for k, v := range oauthMap {
		oauthArray = append(oauthArray, k + "=\"" + Encode(v) + "\"")
	}
	oauth = strings.Join(oauthArray, ", ")
	return "OAuth " + oauth
}

func GetPublicTimeline(c appengine.Context) ([]TweetTw, os.Error) {
	url := "http://api.twitter.com/statuses/public_timeline.json"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", OAuthHeader(c, "GET", url))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("GetPublicTimeline failed to api call: %v", err.String())
		return nil, err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("GetPublicTimeline failed to read result: %v", err.String())
		return nil, err
	}
	tweets := make([]TweetTw, 0, 20)
	json.Unmarshal(jsonVal, &tweets)
	if len(tweets) == 0 {
		//maybe error.
		var errRes ErrorResponse
		json.Unmarshal(jsonVal, &errRes)
		c.Errorf("GetPublicTimeline error message: %v", errRes.Error)
		return nil, os.NewError(errRes.Error)
	}
	for _, tweet := range tweets {
		c.Debugf("GetPublicTimeline %v: %v", tweet.User.Screen_name, tweet.Text)
	}
	return tweets, nil
}

func InvokePublicTimelineStream(c appengine.Context, procTweet func(TweetTw) os.Error) os.Error {
	request, _ := http.NewRequest(
			 "POST",
			 "http://stream.twitter.com/1/statuses/filter.json",
			 bytes.NewBufferString("track=#,＃"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Basic aW1ha2FyYXlhcnU6N2tvcm9iaThva2k=")
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("InvokePublicTimelineStream failed to call api: %v", err.String())
		return err
	}
	const NBUF = 512
	var buf [NBUF]byte
	var str string
	for {
		switch nr, er := response.Body.Read(buf[:]); true {
		case nr < 0:
			c.Errorf("InvokePublicTimelineStream read to api result: %v", er.String())
			os.Exit(1)
		case nr == 0: // EOF
			return nil
		case nr > 0:
			str += string(buf[0:nr])
			tweetsStrings := strings.Split(str, "\n", -1)
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

func GetTrends(c appengine.Context) ([]Trend, os.Error) {
	url := "http://api.twitter.com/1/trends/1118370.json"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", OAuthHeader(c, "GET", url))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("GetTrends failed to api call: %v", err.String())
		return nil, err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("GetTrends failed to read result: %v", err.String())
		return nil, err
	}
	locales := make([]Locale, 0, 1)
	json.Unmarshal(jsonVal, &locales)
	if len(locales) == 0 || len(locales[0].Trends) == 0 {
		//maybe error.
		var errRes ErrorResponse
		json.Unmarshal(jsonVal, &errRes)
		return nil, os.NewError(errRes.Error)
	}
	for _, trend := range locales[0].Trends {
		c.Debugf("GetTrends trend.Name: %v", trend.Name)
	}
	return locales[0].Trends, nil
}

func HomeTest(c appengine.Context) os.Error {
	url := "http://api.twitter.com/statuses/friends_timeline.json"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", OAuthHeader(c, "GET", url))
	c.Debugf("HomeTest Authorization: %v", request.Header.Get("Authorization"))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("HomeTest failed to api call: %v", err.String())
		return err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("HomeTest failed to read result: %v", err.String())
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

func SearchTweetsByHashtag(c appengine.Context, hashtag string) ([]TweetTw, os.Error) {
	var empty []TweetTw
	if len(hashtag) == 0 {
		return empty, nil
	}
	url := "http://search.twitter.com/search.json?rpp=100&q=" + Encode(hashtag)
	c.Debugf("SearchTweetsByHashtag url: %s", url)
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", OAuthHeader(c, "GET", url))
	c.Debugf("HomeTest Authorization: %v", request.Header.Get("Authorization"))
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("SearchTweetsByHashtag failed to api call: %v", err.String())
		return nil, err
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("SearchTweetsByHashtag failed to read result: %v", err.String())
		return nil, err
	}
	var result Result
	json.Unmarshal(jsonVal, &result)
	if len(result.Results) == 0 {
		//maybe error.
		var errRes ErrorResponse
		json.Unmarshal(jsonVal, &errRes)
		return nil, os.NewError(errRes.Error)
	}
//	for _, tweet := range result.Results {
//		c.Debugf("SearchTweetsByHashtag tweet.Text: %v", tweet.Text)
//	}
	return result.Results, nil
}
