package utamaru

import (
	"appengine"
	"appengine/taskqueue"
	"fmt"
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
			if !ContainsMultibyteChar(hashtag) {
				continue
			}
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
	hashtags, err := GetHashtagsFromRss(c)
	if err != nil {
		c.Errorf("RecordRssHashtags failed to GetTrends: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	for _, h := range hashtags {
		fmt.Fprintf(w, "<b>%v</b><br>\n", h.Name)
		if err := SaveHashtag(c, h.Name, 3); err != nil {
			c.Errorf("RecordRssHashtags failed to SaveHashtag: %v", err.String())
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
	}
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
		if !ContainsMultibyteChar(hashtag.Name) {
			c.Infof("not contains multibyte char: %s", hashtag.Name)
			continue
		}
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
	var adminTemplate = template.MustParseFile("templates/admin.html", nil)
	c.Infof("CronAdmin")
	if err := adminTemplate.Execute(w, nil); err != nil {
		c.Errorf("FrontTop failed to merge template: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}
