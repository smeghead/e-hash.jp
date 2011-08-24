package utamaru

import (
	"os"
	"appengine"
	"appengine/datastore"
	"time"
)

type Hashtag struct {
	Name string
	Count int
	Date datastore.Time
	Crawled datastore.Time
}

func SaveHashtag(c appengine.Context, hashtag string) os.Error {
	//search
	h := new(Hashtag)
	key := datastore.NewKey("Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, h); err != nil {
		//insert
		h.Name = hashtag
		h.Count = 0
	}
	// Countは最大5にする。
	if h.Count < 5 {
		h.Count += 1
	}
	h.Date = datastore.SecondsToTime(time.Seconds())

	if _, err := datastore.Put(c, key, h); err != nil {
		c.Errorf("SaveHashtag failed to put: %v", err.String())
		return err
	}

	// 古くてカウントが多いもののカウントを減らす
	length := 10
	q := datastore.NewQuery("Hashtag").Filter("Count >", 2).Order("-Date").Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		return err
	}
	for _, h := range hashtags {
		if h.Count <= 1 {
			continue;
		}
		c.Debugf("SaveHashtag old hashtag decrement: %v", h.Name)
		h.Count -= 1
		if _, err := datastore.Put(c, key, h); err != nil {
			c.Errorf("SaveHashtag failed to put old hashtag decrement: %v", err.String())
			return err
		}
	}
	return nil
}

func UpdateHashtag(c appengine.Context, hashtag string) os.Error {
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

func FindHashtag(c appengine.Context, hashtag string) (Hashtag, os.Error) {
	//search
	h := new(Hashtag)
	key := datastore.NewKey("Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, h); err != nil {
		return *h, err
	}
	return *h, nil
}

func GetHashtags(c appengine.Context, options map[string]interface{}) ([]Hashtag, os.Error) {
	length := options["length"].(int)
	//search
	q := datastore.NewQuery("Hashtag").Order("Crawled").Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		return nil, err
	}
	c.Debugf("len hashtags: %d", len(hashtags))
	return hashtags, nil
}

func GetPublicHashtags(c appengine.Context, options map[string]interface{}) ([]Hashtag, os.Error) {
	length := options["length"].(int)
	noCache := 0
	if options["noCache"] != nil {
		noCache = options["noCache"].(int)
	}
	if noCache == 0 {
		// try to get cache.
		hs, err := CacheGetSubjects(c, options)
		if err == nil {
			// got from memcached.
			c.Debugf("GetPublicHashtags got from memcached")
			return hs, nil
		}
	}
	//search
	q := datastore.NewQuery("Hashtag").Order("-Count").Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		return nil, err
	}
	c.Debugf("len hashtags: %d", len(hashtags))
	if noCache == 0 {
		// cache.
		CacheSetSubjects(c, hashtags, options)
		c.Debugf("GetPublicHashtags subjects cached")
	}
	return hashtags, nil
}
