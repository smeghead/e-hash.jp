package utamaru

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"http"
	"time"
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
	http.HandleFunc("/point_up", PointUpHandler)
	http.HandleFunc("/like", LikeHandler)
	http.HandleFunc("/oauthlike", OauthLikeHandler)

	http.HandleFunc("/cron/admin", CronAdmin)
	http.HandleFunc("/cron/record_hashtags", RecordHashtags)
	http.HandleFunc("/cron/record_trends_hashtags", RecordTrendsHashtags)
	http.HandleFunc("/cron/record_rss_hashtags", RecordRssHashtags)
	http.HandleFunc("/cron/crawle_hashtags", CrawleHashtags)
	http.HandleFunc("/worker/crawle_hashtag", WorkerCrawleHashtagHandler)
	http.HandleFunc("/home_test", HomeTestHandler)
	http.HandleFunc("/get_request_token", GetReqestTokenHandler)
	http.HandleFunc("/callback", GetAccessTokenHandler)
	http.HandleFunc("/admin/migrate_tweet", MigrateTweetHandler)
	http.HandleFunc("/admin/migrate_hashtag", MigrateHashtagHandler)
}

func ErrorPage(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	var errorTemplate = template.MustParseFile("templates/error.html", nil)
	if err := errorTemplate.Execute(w, map[string]interface{}{
				"siteTitle": SiteTitle,
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
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
	SaveRequestToken(c, *requestToken)
	http.Redirect(w, r, "http://twitter.com/oauth/authorize?oauth_token=" + requestToken.OauthToken, 302)
}

func GetAccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("GetAccessTokenHandler")
	if len(r.FormValue("denied")) > 0 {
		//Oauth no thanks.
		url := getCookie(r, "url")
		if len(url) == 0 {
			url = "/"
		}
		http.Redirect(w, r, url, 302)
	}
	oauthToken := r.FormValue("oauth_token")
	oauthVerifier := r.FormValue("oauth_verifier")
	c.Debugf("%s, %s", oauthToken, oauthVerifier)
	requestToken, err := FindRequestToken(c, oauthToken)
	if err != nil {
		c.Errorf("GetAccessTokenHandler failed to find requestToken: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	c.Debugf("requestToken: %s", requestToken.OauthToken)
	user, err := GetAccessToken(c, requestToken, oauthVerifier)
	if err != nil {
		c.Errorf("GetReqestTokenHandler failed to post: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	user.SessionId = GetUniqId(r.RemoteAddr, r.UserAgent)
	if err := SaveUser(c, *user); err != nil {
		c.Errorf("SaveUser failed to save: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	oneYearLater := time.LocalTime()
	oneYearLater.Year += 1
	http.SetCookie(w, &http.Cookie{
		Name: "id",
		Value: user.SessionId,
		Path: "/",
		Expires: *oneYearLater})
	http.Redirect(w, r, "/oauthlike", 302)
}
