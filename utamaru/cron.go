package utamaru

import (
	"appengine"
	"appengine/taskqueue"
	"fmt"
	"strconv"
	"time"
	"http"
	"regexp"
	"template"
)

// Stream API Version.
// func RecordHashtags(w http.ResponseWriter, r *http.Request) {
// 	c := appengine.NewContext(r)
// 	err := InvokePublicTimelineStream(c, func(t TweetTw) os.Error {
// 		reg, err := regexp.Compile("[#＃][^ ;'.,\n]+")
// 		if err != nil {
// 			c.Errorf("RecordHashtags failed to compile regexp: %v", err.String())
// 			http.Error(w, err.String(), http.StatusInternalServerError)
// 			return err
// 		}
// 		matches := reg.FindAllString(t.Text, 5)
// 		for _, hashtag := range matches {
// 			c.Debugf("RecordHashtags hashtag: %v", hashtag)
// 			if err := SaveHashtag(c, hashtag); err != nil {
// 				c.Errorf("RecordHashtags failed to SaveHashtag: %v", err.String())
// 				http.Error(w, err.String(), http.StatusInternalServerError)
// 				return err
// 			}
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		c.Errorf("RecordHashtags failed to InvokePublicTimelineStream: %v", err.String())
// 		http.Error(w, err.String(), http.StatusInternalServerError)
// 	}
// 	return
// }

func RecordHashtags(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	//public timeline
	tweets, err := GetPublicTimeline(c)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	reg, err := regexp.Compile(HashtagRexexp)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	for _, t := range tweets {
		fmt.Fprintf(w, "%v<br>\n", t.Text)
		if !ContainsMultibyteChar(t.Text) {
			continue
		}
		matches := reg.FindAllString(t.Text, 5)
		for _, hashtag := range matches {
			fmt.Fprintf(w, "<b>%v</b><br>\n", hashtag)
			if err := SaveHashtag(c, hashtag, 0); err != nil {
				http.Error(w, err.String(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func RecordTrendsHashtags(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("RecordTrendsHashtags")
	// decrement old hashtag
	if err := DecrementOldHashtags(c, 50); err != nil {
		c.Errorf("RecordTrendsHashtags failed to DecrementOldHashtags: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
	}

	trends, err := GetTrends(c)
	if err != nil {
		c.Errorf("RecordTrendsHashtags failed to GetTrends: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	for _, t := range trends {
		reg, err := regexp.Compile(HashtagRexexp)
		if err != nil {
			c.Errorf("RecordTrendsHashtags regexp compile error: %v", err.String())
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%v<br>\n", t.Name)
		matches := reg.FindAllString(t.Name, 5)
		for _, hashtag := range matches {
			fmt.Fprintf(w, "<b>%v</b><br>\n", hashtag)
			if err := SaveHashtag(c, hashtag, 0); err != nil {
				c.Errorf("RecordTrendsHashtags failed to SaveHashtag: %v", err.String())
				http.Error(w, err.String(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func RecordRssHashtags(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("RecordRssHashtags")
	// decrement old hashtag
	if err := DecrementOldHashtags(c, 50); err != nil {
		c.Errorf("RecordTrendsHashtags failed to DecrementOldHashtags: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
	}

	hashtags, err := GetHashtagsFromRss(c)
	if err != nil {
		c.Errorf("RecordRssHashtags failed to GetTrends: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	c.Debugf("RecordRssHashtags len(hashtag): %d", len(hashtags))
	for _, h := range hashtags {
		fmt.Fprintf(w, "<b>%v</b><br>\n", h.Name)
		c.Debugf("RecordRssHashtags try to save: %s", h.Name)
		if err := SaveHashtag(c, h.Name, 5); err != nil {
			c.Errorf("RecordRssHashtags failed to SaveHashtag: %v", err.String())
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		c.Debugf("RecordRssHashtags saved: %s", h.Name)
	}
	c.Debugf("RecordRssHashtags end .. ok")
	fmt.Fprintf(w, "end<br>\n")
}

func CrawleHashtags(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("CrawleHashtags")
	hashtags, err := GetHashtags(c, map[string]interface{}{
		"length": 3,
	})
	if err != nil {
		c.Errorf("CrawleHashtags failed to retrieve hashtags: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	for _, hashtag := range hashtags {
		c.Debugf("register taskqueue: %s", hashtag.Name)
		t := taskqueue.NewPOSTTask("/worker/crawle_hashtag", map[string][]string{"hashtag": []string{hashtag.Name}})
		if _, err := taskqueue.Add(c, t, ""); err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
	}
}

func CronAdmin(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	hashtag := r.FormValue("hashtag")
	if len(hashtag) > 0 {
		if err := SaveHashtag(c, hashtag, 3); err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
	}
	var adminTemplate, _ = template.ParseFile("templates/admin.html")
	c.Infof("CronAdmin")
	if err := adminTemplate.Execute(w, nil); err != nil {
		c.Errorf("FrontTop failed to merge template: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}

func CronAdminDeleteTweet(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	p := r.FormValue("p")
	page, _ := strconv.Atoi(p)
	if page < 0 {
		page = 0
	}

	if r.Method == "POST" {
		r.FormValue("hashtags") //parseさせるため
		hashtags := r.Form["hashtags"]
		c.Debugf("CronAdminDeleteTweet: hashtags %v", hashtags)
		c.Debugf("CronAdminDeleteTweet: hashtags length %v", len(hashtags))
		for _, h := range(hashtags) {
			c.Debugf("CronAdminDeleteTweet: hashtag %v", h)
			//delete tweets
			if err := DeleteTweetsByHashtag(c, h); err != nil {
				c.Errorf("CronAdminDeleteTweet failed to delete tweets: %v", err)
				ErrorPage(w, "CronAdminDeleteTweet failed to delete tweets", http.StatusInternalServerError)
				return
			}
		}
		url := fmt.Sprintf("/cron/delete_tweets?p=%d&complete=%d", page, time.Seconds())
		c.Debugf("CronAdminDeleteTweet: redirect to %s", url)
		http.Redirect(w, r, url, 302)
		return
	}
	hashtags, err := GetHashtags(c, map[string]interface{}{
		"page": page,
		"length": 20,
		"order": "View",
	})
	if err != nil {
		c.Errorf("FrontHashtags failed to search hashtags: %v", err)
		ErrorPage(w, "お探しのページが見付かりませんでした。", http.StatusNotFound)
		return
	}
	var deleteTweetsTemplate, _ = template.ParseFile("templates/delete_tweets.html")
	c.Infof("CronAdmin")
	if err := deleteTweetsTemplate.Execute(w, map[string]interface{}{
				"prev": page - 1,
				"next": page + 1,
				"hashtags": hashtags,
			}); err != nil {
		c.Errorf("FrontTop failed to merge template: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}
