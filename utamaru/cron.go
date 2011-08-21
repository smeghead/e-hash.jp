package utamaru

import (
	"os"
	"appengine"
	"fmt"
	"http"
	"regexp"
)

func RecordHashtags(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := InvokePublicTimelineStream(c, func(t TweetTw) os.Error {
		reg, err := regexp.Compile("[#＃][^ ;'.,\n]+")
		if err != nil {
			c.Errorf("RecordHashtags failed to compile regexp: %v", err.String())
			http.Error(w, err.String(), http.StatusInternalServerError)
			return err
		}
		matches := reg.FindAllString(t.Text, 5)
		for _, hashtag := range matches {
			c.Debugf("RecordHashtags hashtag: %v", hashtag)
			if err := SaveHashtag(c, hashtag); err != nil {
				c.Errorf("RecordHashtags failed to SaveHashtag: %v", err.String())
				http.Error(w, err.String(), http.StatusInternalServerError)
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.Errorf("RecordHashtags failed to InvokePublicTimelineStream: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
	return
	//public timeline
//	tweets, err := GetPublicTimeline()
//	if err != nil {
//		http.Error(w, err.String(), http.StatusInternalServerError)
//		return
//	}
//	for _, t := range tweets {
//		reg, err := regexp.Compile("[#＃][^ ;'.,]+")
//		if err != nil {
//			http.Error(w, err.String(), http.StatusInternalServerError)
//			return
//		}
//		fmt.Fprintf(w, "%v<br>\n", t.Text)
//		matches := reg.FindAllString(t.Text, 5)
//		for _, hashtag := range matches {
//			fmt.Fprintf(w, "<b>%v</b><br>\n", hashtag)
//			if err := SaveHashtag(c, hashtag); err != nil {
//				http.Error(w, err.String(), http.StatusInternalServerError)
//				return
//			}
//		}
//	}
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
		reg, err := regexp.Compile("^[#＃].+")
		if err != nil {
			c.Errorf("RecordTrendsHashtags regexp compile error: %v", err.String())
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%v<br>\n", t.Name)
		matches := reg.FindAllString(t.Name, 5)
		for _, hashtag := range matches {
			fmt.Fprintf(w, "<b>%v</b><br>\n", hashtag)
			if err := SaveHashtag(c, hashtag); err != nil {
				c.Errorf("RecordTrendsHashtags failed to SaveHashtag: %v", err.String())
				http.Error(w, err.String(), http.StatusInternalServerError)
				return
			}
		}
	}
}

