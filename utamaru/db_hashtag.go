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
	h.Count += 1
	h.Date = datastore.SecondsToTime(time.Seconds())

	if _, err := datastore.Put(c, key, h); err != nil {
		c.Errorf("SaveHashtag failed to put: %v", err.String())
		return err
	}
	return nil
}

func UpdateHashtag(c appengine.Context, hashtag string) os.Error {
	//search
	h := new(Hashtag)
	key := datastore.NewKey("Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, h); err != nil {
		return nil
	}
	h.Crawled = datastore.SecondsToTime(time.Seconds())

	if _, err := datastore.Put(c, key, h); err != nil {
		c.Errorf("UpdateHashtag failed to put: %v", err.String())
		return err
	}
	return nil
}

func GetHashtags(c appengine.Context) ([]Hashtag, os.Error) {
	//search
	q := datastore.NewQuery("Hashtag").Order("Crawled").Limit(10)
	hashtags := make([]Hashtag, 0, 10)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		return nil, err
	}
	c.Debugf("len hashtags: %d", len(hashtags))
	return hashtags, nil
}

