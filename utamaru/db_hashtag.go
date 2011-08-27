package utamaru

import (
	"os"
	"appengine"
	"appengine/datastore"
	"time"
	"rand"
)

type Hashtag struct {
	Name string
	Count int
	View int
	LastStatusId string
	Date datastore.Time
	Crawled datastore.Time
}

func(h *Hashtag) Valid() bool {
	if !ContainsMultibyteChar(h.Name) {
		return false
	}
	if len(h.Name) <= 5 {
		return false
	}
	return true
}

func SaveHashtag(c appengine.Context, hashtag string, count int) os.Error {
	c.Debugf("SaveHashtag hashtag: %s (%d)", hashtag, count)
	//search
	h := new(Hashtag)
	key := datastore.NewKey("Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, h); err != nil {
		//insert
		h.Name = hashtag
		h.Count = count
		if !h.Valid() {
			c.Warningf("SaveHashtag invalid hashtag. ignoring... hashtag: %s", h.Name)
			return nil
		}
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

	c.Debugf("SaveHashtag ok")
	return nil
}



func DecrementOldHashtags(c appengine.Context, length int) os.Error {
	// 古くてカウントが多いもののカウントを減らす
	q := datastore.NewQuery("Hashtag").Order("Date").Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		c.Errorf("SaveHashtag failed to search hashtags for decrement")
		return err
	}
	c.Debugf("SaveHashtag got old hashtags len: %d", len(hashtags))
	r := rand.New(rand.NewSource(123))
	for _, h := range hashtags {
		if h.Count <= 3 {
			continue;
		}
		c.Debugf("SaveHashtag old hashtag decrement before: %v (%d)", h.Name, h.Count)
		h.Count = (r.Int() % 3) + 1 // 1 or 2 or 3
		c.Debugf("SaveHashtag old hashtag decrement after : %v (%d)", h.Name, h.Count)
		key := datastore.NewKey("Hashtag", h.Name, 0, nil)
		if _, err := datastore.Put(c, key, &h); err != nil {
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
		c.Errorf("UpdateHashtag failed to get: %v", err.String())
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
	var h Hashtag
	key := datastore.NewKey("Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, &h); err != nil {
		c.Errorf("FindHashtag failed to get: %v", err.String())
		return h, err
	}
	return h, nil
}

func ViewHashtag(c appengine.Context, hashtag Hashtag) os.Error {
	hashtag.View += 1

	key := datastore.NewKey("Hashtag", hashtag.Name, 0, nil)
	if _, err := datastore.Put(c, key, &hashtag); err != nil {
		c.Errorf("ViewHashtag failed to put: %v", err.String())
		return err
	}
	return nil
}

func GetHashtags(c appengine.Context, options map[string]interface{}) ([]Hashtag, os.Error) {
	length := options["length"].(int)
	order := "Crawled"
	if options["order"] != nil {
		order = options["order"].(string)
	}
	//search
	q := datastore.NewQuery("Hashtag").Order(order).Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		c.Errorf("GetHashtags failed to get: %v", err.String())
		return nil, err
	}
	c.Debugf("len hashtags: %d", len(hashtags))
	return hashtags, nil
}

func GetPublicHashtags(c appengine.Context, options map[string]interface{}) ([]Hashtag, os.Error) {
	length := options["length"].(int)
	order := "-Count"
	if options["order"] != nil {
		order = options["order"].(string)
	}
	c.Debugf("GetPublicHashtags order: %s", order)
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
	q := datastore.NewQuery("Hashtag").Order(order).Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		c.Errorf("GetPublicHashtags failed to get: %v", err.String())
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

func MigrateHashtag(c appengine.Context, offset, length int) os.Error {
	countQ := datastore.NewQuery("Hashtag")
	count, _ := countQ.Count(c)
	c.Debugf("MigrateHashtag Count: %d", count)

	// 追加カラムの反映
	q := datastore.NewQuery("Hashtag").Order("Date").Offset(offset).Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		c.Errorf("MigrateHashtag failed to search hashtags for decrement")
		return err
	}
	c.Debugf("MigrateHashtag got old hashtags len: %d", len(hashtags))
	for _, hashtag := range hashtags {
		c.Debugf("MigrateHashtag old hashtag : %v", hashtag.Name)
		key := datastore.NewKey("Hashtag", hashtag.Name, 0, nil)
		if _, err := datastore.Put(c, key, &hashtag); err != nil {
			c.Errorf("MigrateHashtag failed to put old hashtag decrement: %v", err.String())
			return err
		}
	}
	return nil
}
