package utamaru

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"http"
)

type Greeting struct {
	Author string
	Content string
	Date datastore.Time
}

func init() {
	http.HandleFunc("/", FrontTop)
	http.HandleFunc("/s/", FrontSubject)
	http.HandleFunc("/cron/record_hashtags", RecordHashtags)
	http.HandleFunc("/cron/record_trends_hashtags", RecordTrendsHashtags)
	http.HandleFunc("/cron/crawle_hashtags", CrawleHashtags)
	http.HandleFunc("/worker/crawle_hashtag", WorkerCrawleHashtagHandler)
	http.HandleFunc("/home_test", HomeTestHandler)
}

func HomeTestHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("HomeTestHandler")
	err := HomeTest(c)
	if err != nil {
		c.Errorf("HomeTestHandler failed to post: %v", err.String())
	}
	fmt.Fprint(w, "end");
}
