package utamaru

// worker for queue tasks.
import (
	"appengine"
	"http"
)


//  引数のハッシュタグを、twitter api search で検索する。
//  取れるだけ取る。
//  それを、Tweetに登録する
func WorkerCrawleHashtagHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("WorkerCrawleHashtagHandler")
	hashtag := r.FormValue("hashtag")
	c.Debugf("WorkerCrawleHashtagHandler hashtag: %v", hashtag)
	tweets, err := SearchTweetsByHashtag(c, hashtag)
	if err != nil {
		c.Errorf("WorkerCrawleHashtagHandler failed to crawle by hashtag: %v", err.String())
		// if this return error, loop execution call Twitter API. so, over the limit.
		//http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if err := SaveTweets(c, tweets, hashtag); err != nil {
		c.Errorf("WorkerCrawleHashtagHandler failed to SaveTweets: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if err := UpdateHashtag(c, hashtag); err != nil {
		c.Errorf("WorkerCrawleHashtagHandler failed to UpdateHashtag: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}

