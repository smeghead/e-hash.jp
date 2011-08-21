
package utamaru

import (
	"appengine"
//	"appengine/datastore"
//	"appengine/user"
//	"fmt"
	"http"
)

/**
  worker for queue tasks.
 */

/**
  引数のハッシュタグを、twitter api search で検索する。
  取れるだけ取る。
  それを、Tweetに登録する
 */
func WorkerCrawleHashtagHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("WorkerCrawleHashtagHandler")
	hashtag := r.FormValue("hashtag")
	c.Debugf("WorkerCrawleHashtagHandler hashtag: %v", hashtag)
	tweets, err := SearchTweetsByHashtag(c, hashtag)
	if err != nil {
		c.Errorf("WorkerCrawleHashtagHandler failed to crawle by hashtag: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if err := SaveTweets(c, tweets, hashtag); err != nil {
		c.Errorf("WorkerCrawleHashtagHandler failed to SaveTweets: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

//	q := datastore.NewQuery("Hashtag").Order("-Count").Limit(10)
//	hashtags := make([]Hashtag, 0, 10)
//	if _, err := q.GetAll(c, &hashtags); err != nil {
//		http.Error(w, err.String(), http.StatusInternalServerError)
//		return
//	}
//	if err := guestbookTemplate.Execute(w, hashtags); err != nil {
//		http.Error(w, err.String(), http.StatusInternalServerError)
//		return
//	}
}

