package utamaru

import (
	"bytes"
	"json"
	"http"
	"appengine"
	"appengine/urlfetch"
	"io/ioutil"
)


type UrlResult struct {
	Id string
}

func ShorterUrl(c appengine.Context, longUrl string) string {
	c.Debugf("ShorterUrl long url: %s", longUrl)
	conf, err := GetTwitterConf(c)
	if err != nil {
		c.Errorf("ShorterUrl failed to load TwitterConf: %v", err.String())
		return ""
	}

	url := "https://www.googleapis.com/urlshortener/v1/url?key=" + conf.GoogleShortUrlApiKey 
	jsonInput, err := json.Marshal(map[string]string{"longUrl": longUrl})
	if err != nil {
		c.Errorf("ShorterUrl failed to marshal json: %v", err.String())
		return ""
	}
	c.Debugf("body: %s", string(jsonInput))
	body := bytes.NewBuffer(jsonInput)
	request, _ := http.NewRequest("POST", url, body)
	request.Header.Set("Content-Type", "application/json")
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("ShorterUrl failed to api call: %v", err.String())
		return ""
	}
	jsonVal, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c.Errorf("ShorterUrl failed to read result: %v", err.String())
		return ""
	}
	c.Debugf("ShorterUrl response: %v", string(jsonVal))
	var result UrlResult
	json.Unmarshal(jsonVal, &result)
	c.Debugf("ShorterUrl %s", result.Id)
	return result.Id
}

