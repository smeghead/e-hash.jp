package utamaru

import (
	"os"
	"json"
	"appengine"
	"appengine/memcache"
)

var cacheLifetime = 60 * 1 // 1 minutes.

func CacheSetSubjects(c appengine.Context, subjects []Hashtag, options map[string]interface{}) {
	optionsVal, _ := json.Marshal(options)
	jsonVal, err := json.Marshal(subjects)
	if err != nil {
		c.Debugf("CacheSetSubjects marshal failed.")
		return
	}
	item := &memcache.Item{
		Key:   "TopSubjects|" + string(optionsVal),
		Value: jsonVal,
		Expiration: int32(cacheLifetime),
	}
	c.Debugf("CacheSetSubjects Key: %s Expiration: %d", item.Key, item.Expiration)
	if err := memcache.Add(c, item); err == memcache.ErrNotStored {
		c.Debugf("CacheSetSubjects item with key %q already exists", item.Key)
		if err := memcache.Set(c, item); err != nil {
			c.Errorf("CacheSetSubjects put failed.", item.Key)
		}
	} else if err != nil {
		c.Debugf("CacheSetSubjects error adding item: %v", err)
	}
}

func CacheGetSubjects(c appengine.Context, options map[string]interface{}) ([]Hashtag, os.Error) {
	optionsVal, _ := json.Marshal(options)
	c.Debugf("CacheSetSubjects Key: %s", "TopSubjects|" + string(optionsVal))
	item, err := memcache.Get(c, "TopSubjects|" + string(optionsVal))
	emptySubjects := make([]Hashtag, 0, 0)
	if err == memcache.ErrCacheMiss {
		c.Debugf("CacheGetSubjects item not in the cache")
		return emptySubjects, err
	} else if err != nil {
		c.Debugf("CacheGetSubjects error getting item: %v", err)
		return emptySubjects, err
	}
	c.Debugf("CacheGetSubjects Key: %s Expiration: %d", item.Key, item.Expiration)
	//c.Debugf("CacheGetSubjects %v", string(item.Value))
	subjects := make([]Hashtag, 0, 20)
	json.Unmarshal(item.Value, &subjects)
	return subjects, nil
}
