package utamaru

import (
	"os"
	"fmt"
	"appengine"
	"http"
	"template"
	"strconv"
	"encoding/base64"
)

type HashtagListElement struct {
	Hashtag Hashtag
	Tweets []Tweet
}

func getCookie(r *http.Request, name string) string {
	for _, c := range r.Cookie {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}
func getSessionId(c appengine.Context, r *http.Request) string {
	return getCookie(r, "id")
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
		c.Errorf("getCommonMap failed to retrieve hashtags: %v", err.String())
		return commonMap, err
	}
	commonMap["Common_HashtagsHot"] = hashtagsHot

	hashtagsNew, err := GetPublicHashtags(c, map[string]interface{}{
		"length": 10,
		"order": "-Date",
	})
	if err != nil {
		c.Errorf("getCommonMap failed to retrieve hashtags: %v", err.String())
		return commonMap, err
	}
	commonMap["Common_HashtagsNew"] = hashtagsNew
	
	return commonMap, nil
}

var topTemplate = template.MustParseFile("templates/index.html", nil)
func FrontTop(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	c.Debugf("session id: %s", getSessionId(c, r))
	resultMap, err := getCommonMap(c)
	if err != nil {
		c.Errorf("FrontTop failed to retrieve resultMap: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}

	hashtagsForTop, err := GetPublicHashtags(c, map[string]interface{}{
		"length": 10,
		"order": "random",
	})
	if err != nil {
		c.Errorf("FrontTop failed to retrieve hashtags: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}

	topCount := 5
	hles := make([]HashtagListElement, 0, topCount)
	for _, h := range hashtagsForTop {
		c.Debugf("FrontTop hashtag: %v", h.Name)
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
		c.Errorf("FrontSubject failed to retrieve resultMap: %v", err.String())
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
	// debug
	c.Debugf("FrontSubject user debug")
	for _, t := range tweets {
		for _, u := range t.Users {
			c.Debugf("FrontSubject user: %v", u)
		}
	}

	resultMap["elements"] = hle
	if err := subjectTemplate.Execute(w, resultMap); err != nil {
		c.Errorf("FrontSubject failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
	c.Debugf("FrontSubject end")
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
	c.Debugf("PointUpHandler")

	if r.Method != "POST" {
		c.Errorf("PointUpHandler method not supported.")
		ErrorPage(w, "ポイントアップに失敗しました。", http.StatusInternalServerError)
		return
	}
	key := r.FormValue("key")
	pointType := r.FormValue("type")

	err := PointUpTweet(c, key, pointType)
	if err != nil {
		c.Errorf("PointUpHandler failed to point up. : %v", err)
		ErrorPage(w, "ポイントアップに失敗しました。", http.StatusInternalServerError)
		return
	}
}

func OauthLikeHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("OauthLikeHandler")

	keyBytes, _ := base64.StdEncoding.DecodeString(getCookie(r, "key"))
	key := string(keyBytes)
	url := getCookie(r, "url")
	if len(url) == 0 {
		url = "/"
	}
	c.Debugf("OauthLikeHandler key: %s url: %s", key, url)
	sessionId := getSessionId(c, r)
	if sessionId == "" {
		c.Infof("no auth information. redirect to oauth confirmation.")
		http.Redirect(w, r, url, 302)
		return
	}
	user, err := FindUser(c, sessionId)
	if err != nil {
		c.Infof("OauthLikeHandler failed to find user. %v", err)
		http.Redirect(w, r, url, 302)
		return
	}
	if user.ScreenName == "" {
		c.Infof("OauthLikeHandler failed to find user. empty user.")
		http.Redirect(w, r, url, 302)
		return
	}
	c.Debugf("OauthLikeHandler user: %v", user)

	err = LikeTweet(c, key, user)
	if err != nil {
		c.Errorf("OauthLikeHandler failed to point up. : %v", err)
		ErrorPage(w, "Likeに失敗しました。", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, 302)
}

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("LikeHandler")

	if r.Method != "POST" {
		c.Errorf("LikeHandler method not supported.")
		ErrorPage(w, "Likeに失敗しました。", http.StatusInternalServerError)
		return
	}
	// パラメータの保存 Oauthの後にリダイレクトするため。
	key := r.FormValue("key")
	c.Debugf("LikeHandler key: %s", key)
	http.SetCookie(w, &http.Cookie{
		Name: "key",
		Value: base64.StdEncoding.EncodeToString([]byte(key)),
		Path: "/",
	})
	url := r.FormValue("url")
	c.Debugf("LikeHandler url: %s", url)
	http.SetCookie(w, &http.Cookie{
		Name: "url",
		Value: url,
		Path: "/",
	})
	sessionId := getSessionId(c, r)
	if sessionId == "" {
		c.Infof("no auth information. redirect to oauth confirmation.")
		fmt.Fprint(w, "needs_oauth")
		return
	}
	user, err := FindUser(c, sessionId)
	if err != nil {
		c.Infof("LikeHandler failed to find user. %v", err)
		fmt.Fprint(w, "needs_oauth")
		return
	}
	if user.ScreenName == "" {
		c.Infof("LikeHandler failed to find user. empty user.")
		fmt.Fprint(w, "needs_oauth")
		return
	}
	c.Debugf("LikeHandler user: %v", user)


	err = LikeTweet(c, key, user)
	if err != nil {
		c.Errorf("LikeHandler failed to point up. : %v", err)
		ErrorPage(w, "Likeに失敗しました。", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, user.ScreenName)
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
