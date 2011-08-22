package utamaru

import (
	"os"
	"appengine"
	"appengine/datastore"
	"time"
)

type Tweet struct {
	Id_Str string
	Screen_name string
	UserId_Str string
	Text string
	Profile_Image_Url string
	Created_At datastore.Time
	Hashtag string
}

func (t *Tweet) String() string {
	return t.Hashtag + ":" + t.Id_Str
}

func NewTweet(tw TweetTw) Tweet {
	var t Tweet
	t.Id_Str = tw.Id_Str
	t.Screen_name = tw.User.Screen_name
	t.UserId_Str = tw.User.Id_Str
	t.Text = tw.Text
	t.Profile_Image_Url = tw.Profile_Image_Url
	createAtTime, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", tw.Created_At)
	t.Created_At = datastore.SecondsToTime(createAtTime.Seconds())
	return t
}

func SaveTweets(c appengine.Context, tweets []TweetTw, hashtag string) os.Error {
	for _, tweet := range tweets {
		if !ContainsMultibyteChar(hashtag) {
			c.Infof("not contains multibyte char: %s", hashtag)
			continue
		}
		t := NewTweet(tweet)
		t.Hashtag = hashtag
		key := datastore.NewKey("Tweet", t.String(), 0, nil)
		c.Debugf("key: %v", key)

		if _, err := datastore.Put(c, key, &t); err != nil {
			c.Errorf("SaveTweets failed to put: %v", err.String())
			return err
		}
	}
	return nil
}

func ContainsMultibyteChar(s string) bool {
	for _, c := range s {
		if c > 128 {
			return true
		}
	}
	return false
}
