package utamaru

import (
	"errors"
	"appengine"
	"appengine/datastore"
	"time"
	"strings"
//	"regexp"
)

type TweetLikeUser struct {
	TweetKey string
	LikeType string
	UserId string
	ScreenName string
	Created_At time.Time
}
type Tweet struct {
	Id_Str string
	Screen_name string
	UserId_Str string
	Text string
	Profile_Image_Url string
	Created_At time.Time
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
	createAtTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", tw.Created_At)
	if err != nil {
		createAtTime = time.Now().Local()
	}
	t.Created_At = createAtTime
	return t
}

func CopyTweet(newT, old Tweet) Tweet {
	old.Screen_name = newT.Screen_name
	old.UserId_Str = newT.UserId_Str
	old.Text = newT.Text
	old.Profile_Image_Url = newT.Profile_Image_Url
	return old
}

func SaveTweets(c appengine.Context, tweets []TweetTw, hashtag string) error {
	for _, tweet := range tweets {
		if len(strings.Trim(tweet.Text, " 　\n")) == len(hashtag) {
			// ハッシュタグだけのtweetは登録しない
			c.Infof("SaveTweets tweet is only hashtag: %s", hashtag)
			continue
		}
		c.Debugf("SaveTweets tweet : %s", tweet.Text)
		//if len(tweet.To_User_Id_Str) > 0 || tweet.Text[0:4] == "RT @" {
		if tweet.Text[0:4] == "RT @" {
			// RTは、無視する
			c.Infof("SaveTweets tweet is RT: %s", tweet.Text)
			continue
		}
		t := NewTweet(tweet)
		t.Hashtag = hashtag
		key := datastore.NewKey(c, "Tweet", t.String(), 0, nil)

		var old Tweet
		if err := datastore.Get(c, key, &old); err == nil {
			// 既に存在する場合
			t = CopyTweet(t, old)
		}
		if _, err := datastore.Put(c, key, &t); err != nil {
			c.Errorf("SaveTweets failed to put: %v", err)
			return err
		}
	}

	// まったくtweetが無いならhashtagを削除する
	count, err := datastore.NewQuery("Tweet").Filter("Hashtag =", hashtag).Count(c)
	if err != nil {
		c.Errorf("SaveTweets failed to count: %v", err)
		return err
	}

	if count == 0 {
		c.Infof("SaveTweets delete hashtag cas no tweets.")
		DeleteHashtag(c, hashtag);
	}

	if len(tweets) == 0 {
		return nil
	}
	//Hashtag.LastStatusIdの記録
	lastStatusId := tweets[len(tweets) - 1].Id_Str
	if h, err := FindHashtag(c, hashtag); err == nil {
		h.LastStatusId = lastStatusId
		key := datastore.NewKey(c, "Hashtag", hashtag, 0, nil)
		if _, er := datastore.Put(c, key, &h); er != nil {
			c.Errorf("SaveTweets failed to put hashtag: %v", er)
		}
	}
	return nil
}

func SaveTweet(c appengine.Context, t Tweet, hashtag string) error {
	t.Hashtag = hashtag
	key := datastore.NewKey(c, "Tweet", t.String(), 0, nil)

	if _, err := datastore.Put(c, key, &t); err != nil {
		c.Errorf("SaveTweet failed to put: %v", err)
		return err
	}

	return nil
}

func UpdateTweet(c appengine.Context, hashtag string) error {
	//search
	h := new(Hashtag)
	key := datastore.NewKey(c, "Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, h); err != nil {
		return err
	}
	h.Crawled = time.Now()

	if _, err := datastore.Put(c, key, h); err != nil {
		c.Errorf("UpdateHashtag failed to put: %v", err)
		return err
	}
	return nil
}

func PointUpTweet(c appengine.Context, keyString, pointType string, user TwitterUser) error {
	c.Debugf("PointUpTweet start")
	//search
	var tweet Tweet

	c.Debugf("tweet pointType: %s key: %s", pointType, keyString)
	key := datastore.NewKey(c, "Tweet", keyString, 0, nil)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		if err := datastore.Get(c, key, &tweet); err != nil {
			c.Errorf("PointUpTweet failed to get: %v", err)
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
			return errors.New("invalid pointType")
		}

		// ポイントの計算
		tweet.Point += getPoint(c, tweet)

		if _, err := datastore.Put(c, key, &tweet); err != nil {
			c.Errorf("PointUpTweet failed to put: %v", err)
			return err
		}
		return nil
	}, nil)
	if err != nil {
		c.Errorf("Transaction failed: %v", err)
		return err
	}
	// ユーザの記録
	c.Debugf("PointUpTweet user.ScreenName: %v", user)
	if user.ScreenName != "" {
		c.Debugf("PointUpTweet save user")
		tweetLikeUser := TweetLikeUser{
			TweetKey: tweet.Hashtag + ":" + tweet.Id_Str,
			LikeType: pointType,
			UserId: user.Id,
			ScreenName: user.ScreenName,
			Created_At: time.Now(),
		}
		c.Debugf("PointUpTweet user %v", tweetLikeUser)
		key := datastore.NewKey(c, "TweetLikeUser", tweet.Hashtag + ":" + tweet.Id_Str + ":" + user.ScreenName, 0, nil)
		if _, err := datastore.Put(c, key, &tweetLikeUser); err != nil {
			c.Errorf("PointUpTweet TweetLikeUser failed to put: %v", err)
			return err
		}
		c.Debugf("PointUpTweet saved user")
	}
	c.Debugf("PointUpTweet end")
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

func LikeTweet(c appengine.Context, keyString string, user TwitterUser) error {
	var tweet Tweet

	err := datastore.RunInTransaction(c, func(c appengine.Context) error {

		c.Debugf("tweet key: %s", keyString)
		key := datastore.NewKey(c, "Tweet", keyString, 0, nil)
		if err := datastore.Get(c, key, &tweet); err != nil {
			c.Errorf("LikeTweet failed to get: %v", err)
			return err
		}
		tweet.LikeCount += 1

		// ポイントの計算
		tweet.Point += getPoint(c, tweet)

		if _, err := datastore.Put(c, key, &tweet); err != nil {
			c.Errorf("LikeTweet failed to put: %v", err)
			return err
		}

		return nil
	}, nil)
	if err != nil {
		c.Errorf("Transaction failed: %v", err)
		return err
	}

	// ユーザの記録
	key := datastore.NewKey(c, "TweetLikeUser", tweet.Hashtag + ":" + tweet.Id_Str + ":" + user.ScreenName, 0, nil)
	tweetLikeUser := TweetLikeUser{
		TweetKey: tweet.Hashtag + ":" + tweet.Id_Str,
		LikeType: "like",
		UserId: user.Id,
		ScreenName: user.ScreenName,
		Created_At: time.Now(),
	}
	if _, err := datastore.Put(c, key, &tweetLikeUser); err != nil {
		c.Errorf("LikeTweet TweetLikeUser failed to put: %v", err)
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

func GetTweetsByHashtag(c appengine.Context, hashtag string, options map[string]interface{}) ([]Tweet, error) {
	length := options["length"].(int)
	page := 0
	if options["page"] != nil {
		page = options["page"].(int)
	}
	order := "-Point"
	if options["sort"] != nil {
		c.Debugf("GetTweetsByHashtag sort: %v", options["sort"].(string))
		if options["sort"].(string) == "new" {
			order = "-Created_At"
		}
	}
	c.Debugf("GetTweetsByHashtag order: %v", order)
	noCache := 0
	if options["noCache"] != nil {
		noCache = options["noCache"].(int)
	}
	if noCache == 0 {
		// try to get cache.
		hs, err := CacheGetTweetsByHashtag(c, hashtag, options)
		if err == nil {
			// got from memcached.
			c.Debugf("GetTweetsByHashtag got from memcached")
			return hs, nil
		}
	}

	//search
	q := datastore.NewQuery("Tweet").Filter("Hashtag =", hashtag).Order(order).Offset(page * length).Limit(length)
	var tweets []Tweet
	if _, err := q.GetAll(c, &tweets); err != nil {
		c.Errorf("GetTweetsByHashtag failed to get: %v", err)
		return nil, err
	}
	var tweetsResult []Tweet
	for _, tweet := range tweets {
		q := datastore.NewQuery("TweetLikeUser").Filter("TweetKey =", tweet.Hashtag + ":" + tweet.Id_Str).Order("-Created_At").Limit(100)
		if _, err := q.GetAll(c, &tweet.Users); err != nil {
			c.Errorf("GetTweetsByHashtag failed to get users: %v", err)
			return nil, err
		}
		tweetsResult = append(tweetsResult, tweet)
	}
	if noCache == 0 {
		// cache.
		CacheSetTweetsByHashtag(c, hashtag, tweetsResult, options)
		c.Debugf("GetTweetsByHashtag tweets cached")
	}
	return tweetsResult, nil
}

func MigrateTweet(c appengine.Context, offset, length int) error {
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
		key := datastore.NewKey(c, "Tweet", tweet.Hashtag + ":" + tweet.Id_Str, 0, nil)
		if _, err := datastore.Put(c, key, &tweet); err != nil {
			c.Errorf("MigrateTweet failed to put old tweet decrement: %v", err)
			return err
		}
	}
	return nil
}

func DeleteTweetsByHashtag(c appengine.Context, hashtag string) error {
	c.Debugf("DeleteTweetsByHashtag")
	q := datastore.NewQuery("Tweet").Filter("Hashtag =", hashtag).Order("Created_At").Limit(300)
	var tweets []Tweet
	keys, err := q.GetAll(c, &tweets)
	if err != nil {
		c.Errorf("DeleteTweetsByHashtag failed to get: %v", err)
		return err
	}
	c.Debugf("DeleteTweetsByHashtag keys for delete: %v", keys)
	for _, tweet := range tweets {
		c.Infof("DeleteTweetsByHashtag tweet for delete: %v", tweet.Text)
	}
	if err = datastore.DeleteMulti(c, keys); err != nil {
		c.Errorf("DeleteTweetsByHashtag failed to delete tweet: %v", err)
		return err
	}

	//tweetを削除した結果、tweetが無かったら、Hashtagを削除する
	length, err := datastore.NewQuery("Tweet").Filter("Hashtag =", hashtag).Count(c)
	if err != nil {
		c.Errorf("DeleteTweetsByHashtag failed to get count: %v", err)
		return err
	}
	c.Debugf("DeleteTweetsByHashtag tweets length: %d", length)
	if length == 0 {
		c.Infof("DeleteTweetsByHashtag delete Hashtag: %s", hashtag)
		if err := DeleteHashtag(c, hashtag); err != nil {
			c.Errorf("DeleteTweetsByHashtag failed to delete hashtag: %v", err)
			return err
		}
	}
	return nil
}
