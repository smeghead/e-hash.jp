package utamaru

import (
	"os"
	"appengine"
	"http"
	"template"
	"strconv"
)

type HashtagListElement struct {
	Hashtag Hashtag
	Tweets []Tweet
}

func getCommonMap(c appengine.Context) (map[string]interface{}, os.Error) {
	commonMap := make(map[string]interface{})

	commonMap["siteTitle"] = "ギミハッシュ.in α"
	hashtagsForTicker, _ := GetPublicHashtags(c, map[string]interface{}{
		"length": 5,
	})
	commonMap["Common_HashtagsForTicker"] = hashtagsForTicker

	hashtagsHot, err := GetPublicHashtags(c, map[string]interface{}{
		"length": 10,
	})
	if err != nil {
		c.Errorf("FrontTop failed to retrieve hashtags: %v", err.String())
		return commonMap, err
	}
	commonMap["Common_HashtagsHot"] = hashtagsHot

	hashtagsNew, err := GetPublicHashtags(c, map[string]interface{}{
		"length": 10,
		"order": "-Date",
	})
	if err != nil {
		c.Errorf("FrontTop failed to retrieve hashtags: %v", err.String())
		return commonMap, err
	}
	commonMap["Common_HashtagsNew"] = hashtagsNew
	
	return commonMap, nil
}

var topTemplate = template.MustParseFile("templates/index.html", nil)
func FrontTop(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	resultMap, err := getCommonMap(c)
	if err != nil {
		c.Errorf("FrontTop failed to retrieve resultMap: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}

	topCount := 5
	hles := make([]HashtagListElement, 0, topCount)
	for _, h := range resultMap["Common_HashtagsHot"].([]Hashtag) {
		tweets, err := GetTweetsByHashtag(c, h.Name, map[string]interface{}{
			"length": 2,
		})
		if err != nil {
			c.Errorf("FrontTop failed to retrieve tweets: %v", err.String())
			ErrorPage(w, err.String(), http.StatusInternalServerError)
		}
		if len(tweets) == 0 {
			continue
		}
		e := HashtagListElement{h, tweets}
		hles = append(hles, e)
		if len(hles) >= topCount {
			break
		}
	}
	resultMap["elements"] = hles

	if err := topTemplate.Execute(w, resultMap); err != nil {
		c.Errorf("FrontTop failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}

var subjectTemplate = template.MustParseFile("templates/subject.html", nil)
func FrontSubject(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	path := r.URL.Path
	if len(path) < 4 {
		c.Errorf("FrontSubject 404: %v", path)
		ErrorPage(w, "お探しのページが見付かりませんでした。", http.StatusNotFound)
		return
	}
	hashtag := "#" + path[3:]
	c.Debugf("FrontSubject hashtag: %s", hashtag)

	resultMap, err := getCommonMap(c)
	if err != nil {
		c.Errorf("FrontTop failed to retrieve resultMap: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}

	h, err := FindHashtag(c, hashtag)
	if err != nil {
		c.Errorf("FrontSubject 404: %v", path)
		ErrorPage(w, "お探しのページが見付かりませんでした。", http.StatusNotFound)
		return
	} else {
		ViewHashtag(c, h)
	}

	var tweets []Tweet
	tweets, err = GetTweetsByHashtag(c, hashtag, map[string]interface{}{
		"length": 20,
	})
	if err != nil {
		c.Errorf("FrontSubject failed to retrieve tweets: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
	}
	hle := HashtagListElement{h, tweets}

	resultMap["elements"] = hle
	if err := subjectTemplate.Execute(w, resultMap); err != nil {
		c.Errorf("FrontSubject failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}

func FrontHashtags(w http.ResponseWriter, r *http.Request) {
	var hashtagsTemplate = template.MustParseFile("templates/hashtags.html", nil)
	c := appengine.NewContext(r)

	resultMap, err := getCommonMap(c)
	if err != nil {
		c.Errorf("FrontHashtags failed to retrieve resultMap: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}

	hashtags, err := GetHashtags(c, map[string]interface{}{
		"length": 200,
	})
	if err != nil {
		c.Errorf("FrontHashtags failed to search hashtags: %v", err)
		ErrorPage(w, "お探しのページが見付かりませんでした。", http.StatusNotFound)
		return
	}

	resultMap["hashtags"] = hashtags
	if err := hashtagsTemplate.Execute(w, resultMap); err != nil {
		c.Errorf("FrontHashtags failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}

var aboutTemplate = template.MustParseFile("templates/about.html", nil)
func FrontAbout(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	resultMap, err := getCommonMap(c)
	if err != nil {
		c.Errorf("FrontHashtags failed to retrieve resultMap: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}

	if err := aboutTemplate.Execute(w, resultMap); err != nil {
		c.Errorf("FrontHashtags failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}

func PointUpHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	key := r.FormValue("key")
	pointType := r.FormValue("type")

	err := PointUpTweet(c, key, pointType)
	if err != nil {
		c.Errorf("PointUpHandler failed to point up. : %v", err)
		ErrorPage(w, "ポイントアップに失敗しました。", http.StatusInternalServerError)
		return
	}
}

var subjectMoreTemplate = template.MustParseFile("templates/subject_more.html", nil)
func FrontSubjectMore(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	hashtag := r.FormValue("hashtag")
	pageStr := r.FormValue("page")
	c.Debugf("FrontSubjectMore hashtag: %s", hashtag)
	page, _ := strconv.Atoi(pageStr)
	c.Debugf("FrontSubjectMore page: %d", page)

	var tweets []Tweet
	tweets, err := GetTweetsByHashtag(c, hashtag, map[string]interface{}{
		"length": 20,
		"page": page,
	})
	if err != nil {
		c.Errorf("FrontSubjectMore failed to retrieve tweets: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
	}

	if err := subjectMoreTemplate.Execute(w, map[string]interface{}{
				"Tweets":tweets,
			}); err != nil {
		c.Errorf("FrontSubjectMore failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}
