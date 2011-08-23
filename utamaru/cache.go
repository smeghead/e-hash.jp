package utamaru

import (
	"os"
	"json"
	"appengine"
	"appengine/memcache"
)

var cacheLifetime = 60 * 60

func CacheSetSubjects(c appengine.Context, subjects []Hashtag) {
	jsonVal, err := json.Marshal(subjects)
	if err != nil {
		c.Debugf("CacheSetSubjects marshal failed.")
		return
	}
	item := &memcache.Item{
		Key:   "TopSubjects",
		Value: jsonVal,
	}
	if err := memcache.Add(c, item); err == memcache.ErrNotStored {
		c.Debugf("CacheSetSubjects item with key %q already exists", item.Key)
	} else if err != nil {
		c.Debugf("CacheSetSubjects error adding item: %v", err)
	}
}
func CacheGetSubjects(c appengine.Context) ([]Hashtag, os.Error) {
	item, err := memcache.Get(c, "TopSubjects")
	emptySubjects := make([]Hashtag, 0, 0)
	if err == memcache.ErrCacheMiss {
		c.Debugf("CacheGetSubjects item not in the cache")
		return emptySubjects, err
	} else if err != nil {
		c.Debugf("CacheGetSubjects error getting item: %v", err)
		return emptySubjects, err
	}
	//c.Debugf("CacheGetSubjects %v", string(item.Value))
	subjects := make([]Hashtag, 0, 20)
	json.Unmarshal(item.Value, &subjects)
	return subjects, nil
}
