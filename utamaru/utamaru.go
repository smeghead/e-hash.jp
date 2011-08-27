package utamaru

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"http"
	"template"
	"strconv"
)

type Greeting struct {
	Author string
	Content string
	Date datastore.Time
}

func init() {
	http.HandleFunc("/", FrontTop)
	http.HandleFunc("/s/", FrontSubject)
	http.HandleFunc("/s/more", FrontSubjectMore)
	http.HandleFunc("/hashtags", FrontHashtags)
	http.HandleFunc("/about", FrontAbout)

	http.HandleFunc("/cron/admin", CronAdmin)
	http.HandleFunc("/cron/record_hashtags", RecordHashtags)
	http.HandleFunc("/cron/record_trends_hashtags", RecordTrendsHashtags)
	http.HandleFunc("/cron/record_rss_hashtags", RecordRssHashtags)
	http.HandleFunc("/cron/crawle_hashtags", CrawleHashtags)
	http.HandleFunc("/worker/crawle_hashtag", WorkerCrawleHashtagHandler)
	http.HandleFunc("/point_up", PointUpHandler)
	http.HandleFunc("/home_test", HomeTestHandler)
	http.HandleFunc("/get_request_token", GetReqestTokenHandler)
	http.HandleFunc("/admin/migrate_tweet", MigrateTweetHandler)
	http.HandleFunc("/admin/migrate_hashtag", MigrateHashtagHandler)
}

func ErrorPage(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	var errorTemplate = template.MustParseFile("templates/error.html", nil)
	if err := errorTemplate.Execute(w, map[string]interface{}{
				"ErrorMessage": message,
			}); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
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

func MigrateTweetHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("MigrateTweetHandler")
	offset, _ := strconv.Atoi(r.FormValue("offset"))
	c.Debugf("conv : %d", offset)
	err := MigrateTweet(c, offset, 200)
	if err != nil {
		c.Errorf("MigrateTweetHandler failed to post: %v", err.String())
	}
	fmt.Fprint(w, "end");
}

func MigrateHashtagHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("MigrateHashtagHandler")
	offset, _ := strconv.Atoi(r.FormValue("offset"))
	c.Debugf("conv : %d", offset)
	err := MigrateHashtag(c, offset, 200)
	if err != nil {
		c.Errorf("MigrateHashtagHandler failed to post: %v", err.String())
	}
	fmt.Fprint(w, "end");
}

func GetReqestTokenHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("GetReqestTokenHandler")
	requestToken, err := GetRequestToken(c)
	if err != nil {
		c.Errorf("GetReqestTokenHandler failed to post: %v", err.String())
	}
	fmt.Fprintf(w, "request token: %v", requestToken);
}
