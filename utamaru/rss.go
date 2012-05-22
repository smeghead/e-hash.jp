package utamaru

import (
	"io/ioutil"
	"encoding/xml"
	"net/http"
	"appengine"
	"appengine/urlfetch"
)

type Item struct {
	Title string
}

type Channel struct {
	Item []Item
}

type Rss struct {
	Channel Channel
}

func GetHashtagsFromRss(c appengine.Context) ([]Hashtag, error) {
	url := "http://buzztter.com/ja/rss"
	request, _ := http.NewRequest("GET", url, nil)
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("GetHashtagsFromRss failed to api call: %v", err)
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.Errorf("GetHashtagsFromRss failed to get body : %v", err)
		return nil, err
	}
	var rss Rss
	xml.Unmarshal(body, &rss)
	hashtags := make([]Hashtag, 0, 10)
	for _, item := range rss.Channel.Item {
		if item.Title[0:1] == "#" {
			c.Debugf("GetHashtagsFromRss hit hashtag. %s", item.Title)
			var h Hashtag
			h.Name = item.Title
			if h.Valid() {
				hashtags = append(hashtags, h)
			}
		}
	}
	return hashtags, nil
}

