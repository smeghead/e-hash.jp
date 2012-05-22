package utamaru

import (
	"math/big"
	"appengine"
	"appengine/datastore"
	"time"
	"crypto/rand"
)

type Hashtag struct {
	Name string
	Count int
	View int
	LastStatusId string
	Date time.Time
	Crawled time.Time
}

func(h *Hashtag) Valid() bool {
	if !ContainsMultibyteChar(h.Name) {
		return false
	}
	if len(h.Name) < 1 + 3 * 5 { //バイト数
		return false
	}
	return true
}

func SaveHashtag(c appengine.Context, hashtag string, count int) error {
	c.Debugf("SaveHashtag hashtag: %s (%d)", hashtag, count)
	//search
	h := new(Hashtag)
	key := datastore.NewKey(c, "Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, h); err != nil {
		//insert
		h.Name = hashtag
		h.Count = count
	}
	// Countは最大5にする。
	if h.Count < 5 {
		h.Count += 1
	}
	h.Date = time.Now()
	h.Crawled = time.Unix(0, 0)

	c.Debugf("SaveHashtag Name: %s", h.Name)
	if !h.Valid() {
		c.Infof("SaveHashtag %s is invalid. ignore.", h.Name)
		return nil
	}

	if _, err := datastore.Put(c, key, h); err != nil {
		c.Errorf("SaveHashtag failed to put: %v", err)
		return err
	}

	c.Debugf("SaveHashtag ok")
	return nil
}



func DecrementOldHashtags(c appengine.Context, length int) error {
	// 古くてカウントが多いもののカウントを減らす
	q := datastore.NewQuery("Hashtag").Order("-Count").Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		c.Errorf("DecrementOldHashtags failed to search hashtags for decrement")
		return err
	}
	c.Infof("DecrementOldHashtags got old hashtags len: %d", len(hashtags))
	for _, h := range hashtags {
//		if h.Count <= 4 {
//			continue;
//		}
		c.Debugf("DecrementOldHashtags old hashtag decrement before: %v (%d)", h.Name, h.Count)
		newInt, _ := rand.Int(rand.Reader, big.NewInt(2)) // 1 or 2 or 3
		h.Count = int(newInt.Int64()) + 1
		if h.Count > 5 {
			h.Count = 5
		}
		if h.Count < 1 {
			h.Count = 1
		}
		c.Infof("DecrementOldHashtags old hashtag decrement after : %v (%d)", h.Name, h.Count)
		key := datastore.NewKey(c, "Hashtag", h.Name, 0, nil)
		if _, err := datastore.Put(c, key, &h); err != nil {
			c.Errorf("DecrementOldHashtags failed to put old hashtag decrement: %v", err)
			return err
		}
	}
	return nil
}

func UpdateHashtag(c appengine.Context, hashtag string) error {
	//search
	h := new(Hashtag)
	key := datastore.NewKey(c, "Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, h); err != nil {
		c.Errorf("UpdateHashtag failed to get: %v", err)
		return err
	}
	h.Crawled = time.Now()

	if _, err := datastore.Put(c, key, h); err != nil {
		c.Errorf("UpdateHashtag failed to put: %v", err)
		return err
	}
	return nil
}

func FindHashtag(c appengine.Context, hashtag string) (Hashtag, error) {
	//search
	var h Hashtag
	key := datastore.NewKey(c, "Hashtag", hashtag, 0, nil)

	if err := datastore.Get(c, key, &h); err != nil {
		c.Errorf("FindHashtag failed to get: %v", err)
		return h, err
	}
	return h, nil
}

func DeleteHashtag(c appengine.Context, hashtag string) error {
	//search
	key := datastore.NewKey(c, "Hashtag", hashtag, 0, nil)

	if err := datastore.Delete(c, key); err != nil {
		c.Errorf("DeleteHashtag failed to delete: %v", err)
		return err
	}
	return nil
}

func ViewHashtag(c appengine.Context, hashtag Hashtag) error {
	hashtag.View += 1

	key := datastore.NewKey(c, "Hashtag", hashtag.Name, 0, nil)
	if _, err := datastore.Put(c, key, &hashtag); err != nil {
		c.Errorf("ViewHashtag failed to put: %v", err)
		return err
	}
	return nil
}

func GetHashtags(c appengine.Context, options map[string]interface{}) ([]Hashtag, error) {
	length := options["length"].(int)
	page := 0
	if options["page"] != nil {
		page = options["page"].(int)
	}
	order := "Crawled"
	if options["order"] != nil {
		order = options["order"].(string)
	}
	//search
	q := datastore.NewQuery("Hashtag").Order(order).Offset(length * page).Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		c.Errorf("GetHashtags failed to get: %v", err)
		return nil, err
	}
	c.Debugf("len hashtags: %d", len(hashtags))
	return hashtags, nil
}

func GetPublicHashtags(c appengine.Context, options map[string]interface{}) ([]Hashtag, error) {
	length := options["length"].(int)
	order := "-Count"
	if options["order"] != nil {
		order = options["order"].(string)
		if order == "random" {
			newInt, _ := rand.Int(rand.Reader, big.NewInt(2)) // 1 or 2 or 3
			order = []string{"-Count", "-Date", "Crawled"}[int(newInt.Int64())]
			c.Debugf("GetPublicHashtags order random selected: %s", order)
		}
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
		c.Errorf("GetPublicHashtags failed to get: %v", err)
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

func MigrateHashtag(c appengine.Context, offset, length int) error {
	countQ := datastore.NewQuery("Hashtag")
	count, _ := countQ.Count(c)
	c.Debugf("MigrateHashtag Count: %d", count)

	// 追加カラムの反映
	q := datastore.NewQuery("Hashtag").Order("Date").Offset(offset).Limit(length)
	hashtags := make([]Hashtag, 0, length)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		c.Errorf("MigrateHashtag failed to search hashtags for migrate")
		return err
	}
	c.Debugf("MigrateHashtag got hashtags len: %d", len(hashtags))
	for _, hashtag := range hashtags {
		c.Debugf("MigrateHashtag old hashtag : %v", hashtag.Name)
		key := datastore.NewKey(c, "Hashtag", hashtag.Name, 0, nil)
		if _, err := datastore.Put(c, key, &hashtag); err != nil {
			c.Errorf("MigrateHashtag failed to put hashtag migrate: %v", err)
			return err
		}
	}
	return nil
}
