package utamaru

import (
	"os"
	"appengine"
	"appengine/datastore"
	"time"
	"strings"
//	"regexp"
)

type Tweet struct {
	Id_Str string
	Screen_name string
	UserId_Str string
	Text string
	Profile_Image_Url string
	Created_At datastore.Time
	Hashtag string
	Point int
}

func (t *Tweet) String() string {
	return t.Hashtag + ":" + t.Id_Str
}

func NewTweet(tw TweetTw) Tweet {
	var t Tweet
	t.Id_Str = tw.Id_Str
	t.Screen_name = tw.From_User
	t.UserId_Str = tw.User.Id_Str
	t.Text = tw.Text
//	if len(tw.To_User_Id_Str) > 0 {
//		// Officel RT. pick up original tweet user.
//		t.Id_Str = tw.Id_Str
//		t.UserId_Str = tw.To_User_Id_Str
//		t.Screen_name = tw.To_User
//		reg, err := regexp.Compile("^RT [^:]+: ")
//		if err == nil {
//			t.Text = reg.ReplaceAllString(t.Text, "")
//		}
//	}
	t.Profile_Image_Url = tw.Profile_Image_Url
	createAtTime, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", tw.Created_At)
	t.Created_At = datastore.SecondsToTime(createAtTime.Seconds())
	return t
}

func CopyTweet(newT, old Tweet) Tweet {
	old.Screen_name = newT.Screen_name
	old.UserId_Str = newT.UserId_Str
	old.Text = newT.Text
	old.Profile_Image_Url = newT.Profile_Image_Url
	return old
}

func SaveTweets(c appengine.Context, tweets []TweetTw, hashtag string) os.Error {
	for _, tweet := range tweets {
		if len(strings.Trim(tweet.Text, " 　\n")) == len(hashtag) {
			// ハッシュタグだけのtweetは登録しない
			c.Infof("SaveTweets tweet is only hashtag: %s", hashtag)
			continue
		}
		t := NewTweet(tweet)
		t.Hashtag = hashtag
		key := datastore.NewKey("Tweet", t.String(), 0, nil)
		c.Debugf("SaveTweets key: %v", key)

		var old Tweet
		if err := datastore.Get(c, key, &old); err == nil {
			// 既に存在する場合
			t = CopyTweet(t, old)
			c.Debugf("SaveTweets exists %s", t.Screen_name)
		}
		if _, err := datastore.Put(c, key, &t); err != nil {
			c.Errorf("SaveTweets failed to put: %v", err.String())
			return err
		}
	}
	return nil
}

func ContainsMultibyteChar(s string) bool {
	for _, c := range s {
		if c >= 12353 && c <= 12540 { // range [ぁんァンー]
			return true
		}
		if c >= 19968 && c <= 40869 { // range 漢字
			return true
		}
	}
	return false
}

func GetTweetsByHashtag(c appengine.Context, hashtag string, options map[string]interface{}) ([]Tweet, os.Error) {
	length := options["length"].(int)
	//search
	q := datastore.NewQuery("Tweet").Filter("Hashtag =", hashtag).Order("-Created_At").Limit(length)
	tweets := make([]Tweet, 0, length)
	if _, err := q.GetAll(c, &tweets); err != nil {
		return nil, err
	}
	c.Debugf("len tweets: %d", len(tweets))
	return tweets, nil
}
