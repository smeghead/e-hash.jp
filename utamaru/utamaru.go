package utamaru

import (
	"appengine"
	"fmt"
	"net/http"
	"time"
	"html/template"
	"strconv"
)

type Greeting struct {
	Author string
	Content string
	Date time.Time
}

func init() {
	http.HandleFunc("/", FrontTop)
	http.HandleFunc("/s/", FrontSubject)
	http.HandleFunc("/s/more", FrontSubjectMore)
	http.HandleFunc("/hashtags", FrontHashtags)
	http.HandleFunc("/hashtags_more", FrontHashtagsMore)
	http.HandleFunc("/about", FrontAbout)
	http.HandleFunc("/retweet", RetweetHandler)
	http.HandleFunc("/favorite", FavoriteHandler)
	http.HandleFunc("/point_up", PointUpHandler)
	http.HandleFunc("/post", PostHandler)
	http.HandleFunc("/oauthpost", OauthPostHandler)
	http.HandleFunc("/signout", SignoutHandler)

	http.HandleFunc("/cron/admin", CronAdmin)
	http.HandleFunc("/cron/delete_tweets", CronAdminDeleteTweet)
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
	var errorTemplate, _ = template.ParseFiles("templates/error.html")
	if err := errorTemplate.Execute(w, map[string]interface{}{
				"siteTitle": SiteTitle,
				"ErrorMessage": message,
			}); err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		return
	}
}

func HomeTestHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("HomeTestHandler")
	err := HomeTest(c)
	if err != nil {
		c.Errorf("HomeTestHandler failed to post: %v", err)
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
		c.Errorf("MigrateTweetHandler failed to post: %v", err)
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
		c.Errorf("MigrateHashtagHandler failed to post: %v", err)
	}
	fmt.Fprint(w, "end");
}

func GetReqestTokenHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("GetReqestTokenHandler")
	requestToken, err := GetRequestToken(c)
	if err != nil {
		c.Errorf("GetReqestTokenHandler failed to post: %v", err)
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		return
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
	c.Debugf("passsssss")
	requestToken, err := FindRequestToken(c, oauthToken)
	if err != nil {
		c.Errorf("GetAccessTokenHandler failed to find requestToken: %v", err)
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		return
	}
	c.Debugf("requestToken: %s", requestToken.OauthToken)
	user, err := GetAccessToken(c, requestToken, oauthVerifier)
	if err != nil {
		c.Errorf("GetReqestTokenHandler failed to post: %v", err)
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		return
	}
	user.SessionId = GetUniqId(r.RemoteAddr, r.UserAgent())
	if err := SaveUser(c, *user); err != nil {
		c.Errorf("GetReqestTokenHandler failed to save: %v", err)
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		return
	}
	oneYearLater := time.Now().Local()
	oneYearLater = oneYearLater.AddDate(1, 0,0)
	http.SetCookie(w, &http.Cookie{
		Name: "id",
		Value: user.SessionId,
		Path: "/",
		Expires: oneYearLater,
	})

	oauthType := getCookie(r, "type")
	c.Debugf("GetReqestTokenHandler oauthType: %s", oauthType)
	if oauthType == "like" {
		http.Redirect(w, r, "/oauthlike", 302)
	} else if oauthType == "post" {
		http.Redirect(w, r, "/oauthpost", 302)
	} else {
		c.Warningf("GetReqestTokenHandler unknown oauthType")
		http.Redirect(w, r, "/", 302)
	}
}
