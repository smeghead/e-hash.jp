package utamaru

// worker for queue tasks.
import (
//	"fmt"
	"appengine"
	"net/http"
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
		c.Errorf("WorkerCrawleHashtagHandler failed to crawle by hashtag: %v", err)
		// if this return error, loop execution call Twitter API. so, over the limit.
		//http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if err := SaveTweets(c, tweets, hashtag); err != nil {
		c.Errorf("WorkerCrawleHashtagHandler failed to SaveTweets: %v", err)
		//http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if err := UpdateHashtag(c, hashtag); err != nil {
		c.Errorf("WorkerCrawleHashtagHandler failed to UpdateHashtag: %v", err)
		//http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if len(tweets) > 50 {
//		conf, err := GetTwitterConf(c)
//		if err != nil {
//			c.Errorf("oAuthHeader failed to load TwitterConf: %v", err)
//			return
//		}
		// Post Status.
//		url := conf.Url + "s/" + Encode(hashtag[1:])
//		status := fmt.Sprintf("更新しました。「%s」 %s %s", hashtag[1:], ShorterUrl(c, url), SiteTitle)
//		if err := PostTweet(c, status); err != nil {
//			c.Errorf("FrontTop failed to debug post: %v", err)
//			//ErrorPage(w, err.String(), http.StatusInternalServerError)
//			return
//		}
	}
}

