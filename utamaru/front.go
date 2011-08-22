package utamaru

import (
	"appengine"
	"http"
	"template"
)

type HashtagListElement struct {
	Hashtag Hashtag
	Tweets []Tweet
}


func FrontTop(w http.ResponseWriter, r *http.Request) {
	var topTemplate = template.MustParseFile("templates/index.html", nil)
	c := appengine.NewContext(r)
	hashtags, err := GetPublicHashtags(c, map[string]interface{}{
			"length": 10,
		})
	if err != nil {
		c.Errorf("FrontTop failed to retrieve hashtags: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	hles := make([]HashtagListElement, 0, 10)
	for _, h := range hashtags {
		tweets, err := GetTweetsByHashtag(c, h.Name, map[string]interface{}{
			"length": 3,
		})
		if err != nil {
			c.Errorf("FrontTop failed to retrieve tweets: %v", err.String())
			http.Error(w, err.String(), http.StatusInternalServerError)
		}
		e := HashtagListElement{h, tweets}
		hles = append(hles, e)
	}
	if err := topTemplate.Execute(w, map[string]interface{}{"hashtags":hashtags, "elements": hles}); err != nil {
		c.Errorf("FrontTop failed to merge template: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}

func FrontSubject(w http.ResponseWriter, r *http.Request) {
	var subjectTemplate = template.MustParseFile("templates/subject.html", nil)
	c := appengine.NewContext(r)

	path := r.URL.Path
	if len(path) < 4 {
		c.Errorf("FrontSubject 404: %v", path)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	hashtag := path[3:]
	c.Debugf("FrontSubject hashtag: %s", hashtag)

	h, err := FindHashtag(c, hashtag)
	if err != nil {
		c.Errorf("FrontSubject 404: %v", path)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	var tweets []Tweet
	tweets, err = GetTweetsByHashtag(c, hashtag, map[string]interface{}{
		"length": 50,
	})
	if err != nil {
		c.Errorf("FrontSubject failed to retrieve tweets: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
	hle := HashtagListElement{h, tweets}

	hashtags, err := GetPublicHashtags(c, map[string]interface{}{
			"length": 10,
		})
	if err != nil {
		c.Errorf("FrontTop failed to retrieve hashtags: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	if err := subjectTemplate.Execute(w, map[string]interface{}{"hashtags":hashtags, "elements":hle}); err != nil {
		c.Errorf("FrontSubject failed to merge template: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}
