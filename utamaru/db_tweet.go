package utamaru

import (
	"os"
	"appengine"
	"appengine/datastore"
	"time"
	"strings"
//	"regexp"
)

type TweetLikeUser struct {
	TweetKey string
	UserId string
	ScreenName string
	Created_At datastore.Time
}
type Tweet struct {
	Id_Str string
	Screen_name string
	UserId_Str string
	Text string
	Profile_Image_Url string
	Created_At datastore.Time
	Hashtag string
	Point int
	ReplyCount int
	RetweetCount int
	FavoriteCount int
	ProfileCount int
	LikeCount int
	Users []TweetLikeUser
}

func (t *Tweet) String() string {
	return t.Hashtag + ":" + t.Id_Str
}

func NewTweet(tw TweetTw) Tweet {
	var t Tweet
	t.Id_Str = tw.Id_Str
	t.Screen_name = tw.From_User
	t.UserId_Str = tw.From_User_Id_Str
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
		if len(tweet.To_User_Id_Str) > 0 {
			// RTは、無視する
			c.Infof("SaveTweets tweet is RT: %s", hashtag)
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

	//Hashtag.LastStatusIdの記録
	lastStatusId := tweets[len(tweets) - 1].Id_Str
	if h, err := FindHashtag(c, hashtag); err == nil {
		h.LastStatusId = lastStatusId
		key := datastore.NewKey("Hashtag", hashtag, 0, nil)
		if _, er := datastore.Put(c, key, &h); er != nil {
			c.Errorf("SaveTweets failed to put hashtag: %v", er.String())
		}
	}
	return nil
}

func UpdateTweet(c appengine.Context, hashtag string) os.Error {
	//search
	h := new(Hashtag)
	key := datastore.NewKey("Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, h); err != nil {
		return err
	}
	h.Crawled = datastore.SecondsToTime(time.Seconds())

	if _, err := datastore.Put(c, key, h); err != nil {
		c.Errorf("UpdateHashtag failed to put: %v", err.String())
		return err
	}
	return nil
}

func PointUpTweet(c appengine.Context, keyString, pointType string) os.Error {
	//search
	var tweet Tweet

	c.Debugf("tweet pointType: %s key: %s", pointType, keyString)
	err := datastore.RunInTransaction(c, func(c appengine.Context) os.Error {
		key := datastore.NewKey("Tweet", keyString, 0, nil)
		if err := datastore.Get(c, key, &tweet); err != nil {
			c.Errorf("PointUpTweet failed to get: %v", err.String())
			return err
		}
		switch pointType {
		case "profile":
			tweet.ProfileCount += 1
		case "reply":
			tweet.ReplyCount += 1
		case "retweet":
			tweet.RetweetCount += 1
		case "favorite":
			tweet.FavoriteCount += 1
		default:
			c.Errorf("invalid pointType: %s", pointType)
			return os.NewError("invalid pointType")
		}

		// ポイントの計算
		tweet.Point += getPoint(c, tweet)

		if _, err := datastore.Put(c, key, &tweet); err != nil {
			c.Errorf("PointUpTweet failed to put: %v", err.String())
			return err
		}
		return nil
	})
	if err != nil {
		c.Errorf("Transaction failed: %v", err)
		return err
	}
	return nil
}

func getPoint(c appengine.Context, tweet Tweet) int {
	// ポイントの計算
	return (tweet.LikeCount * 10) +
		(tweet.ProfileCount * 3) +
		(tweet.ReplyCount * 2) +
		(tweet.RetweetCount * 4) +
		(tweet.FavoriteCount * 5)
}

func LikeTweet(c appengine.Context, keyString string, user TwitterUser) os.Error {
	var tweet Tweet

	err := datastore.RunInTransaction(c, func(c appengine.Context) os.Error {

		c.Debugf("tweet key: %s", keyString)
		key := datastore.NewKey("Tweet", keyString, 0, nil)
		if err := datastore.Get(c, key, &tweet); err != nil {
			c.Errorf("LikeTweet failed to get: %v", err.String())
			return err
		}
		tweet.LikeCount += 1

		// ポイントの計算
		tweet.Point += getPoint(c, tweet)

		if _, err := datastore.Put(c, key, &tweet); err != nil {
			c.Errorf("LikeTweet failed to put: %v", err.String())
			return err
		}

		return nil
	})
	if err != nil {
		c.Errorf("Transaction failed: %v", err)
		return err
	}

	// ユーザの記録
	key := datastore.NewKey("TweetLikeUser", tweet.Hashtag + ":" + tweet.Id_Str, 0, nil)
	tweetLikeUser := TweetLikeUser{
		TweetKey: tweet.Hashtag + ":" + tweet.Id_Str,
		UserId: user.Id,
		ScreenName: user.ScreenName,
		Created_At: datastore.SecondsToTime(time.Seconds()),
	}
	if _, err := datastore.Put(c, key, &tweetLikeUser); err != nil {
		c.Errorf("LikeTweet TweetLikeUser failed to put: %v", err.String())
		return err
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
	page := 0
	if options["page"] != nil {
		page = options["page"].(int)
	}
	//search
	q := datastore.NewQuery("Tweet").Filter("Hashtag =", hashtag).Order("-Point").Offset(page * length).Limit(length)
	var tweets []Tweet
	if _, err := q.GetAll(c, &tweets); err != nil {
		c.Errorf("GetTweetsByHashtag failed to get: %v", err.String())
		return nil, err
	}
	var tweetsResult []Tweet
	for _, tweet := range tweets {
		q := datastore.NewQuery("TweetLikeUser").Filter("TweetKey =", tweet.Hashtag + ":" + tweet.Id_Str).Order("-Created_At").Limit(100)
		c.Debugf("TweetLikeUser key: %s %s", tweet.Hashtag + ":" + tweet.Id_Str, tweet.Text)
		if _, err := q.GetAll(c, &tweet.Users); err != nil {
			c.Errorf("GetTweetsByHashtag failed to get users: %v", err.String())
			return nil, err
		}
		c.Debugf("TweetLikeUser length: %d", len(tweet.Users))
		for _, u := range tweet.Users {
			c.Debugf("TweetLikeUser user: %s", u.UserId)
		}
		tweetsResult = append(tweetsResult, tweet)
	}
	return tweetsResult, nil
}

func MigrateTweet(c appengine.Context, offset, length int) os.Error {
	countQ := datastore.NewQuery("Tweet")
	count, _ := countQ.Count(c)
	c.Debugf("MigrateTweet Count: %d", count)

	// 追加カラムの反映
	q := datastore.NewQuery("Tweet").Order("Created_At").Offset(offset).Limit(length)
	tweets := make([]Tweet, 0, length)
	if _, err := q.GetAll(c, &tweets); err != nil {
		c.Errorf("MigrateTweet failed to search tweets for decrement")
		return err
	}
	c.Debugf("MigrateTweet got old tweets len: %d", len(tweets))
	for _, tweet := range tweets {
		c.Debugf("MigrateTweet old tweet : %v", tweet.Text)
		key := datastore.NewKey("Tweet", tweet.Hashtag + ":" + tweet.Id_Str, 0, nil)
		if _, err := datastore.Put(c, key, &tweet); err != nil {
			c.Errorf("MigrateTweet failed to put old tweet decrement: %v", err.String())
			return err
		}
	}
	return nil
}
