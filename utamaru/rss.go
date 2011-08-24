package utamaru

import (
	"os"
	"xml"
	"http"
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

func GetHashtagsFromRss(c appengine.Context) ([]Hashtag, os.Error) {
	url := "http://buzztter.com/ja/rss"
	request, _ := http.NewRequest("GET", url, nil)
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	if err != nil {
		c.Errorf("GetHashtagsFromRss failed to api call: %v", err.String())
		return nil, err
	}
	var rss Rss
	xml.Unmarshal(response.Body, &rss)
	hashtags := make([]Hashtag, 0, 10)
	for _, item := range rss.Channel.Item {
		c.Debugf("GetHashtagsFromRss %v", item.Title)
		if item.Title[0:1] == "#" && ContainsMultibyteChar(item.Title) && len(item.Title) > 5 {
			c.Debugf("GetHashtagsFromRss hit hashtag. %s", item.Title)
			var h Hashtag
			h.Name = item.Title
			hashtags = append(hashtags, h)
		}
	}
	return hashtags, nil
}
