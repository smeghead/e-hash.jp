package utamaru

import (
	"os"
	"fmt"
	"strings"
	"appengine"
	"http"
	"template"
	"strconv"
	"encoding/base64"
)

var SiteTitle = "#イーハッシュJP"

type HashtagListElement struct {
	Hashtag Hashtag
	Tweets []Tweet
}
func isMobile(r *http.Request) bool {
	mobileKeywords := []string{ "Android", "iPhone" }
	for _, keyword := range mobileKeywords {
		if strings.Index(r.UserAgent(), keyword) > -1 {
			return true
		}
	}
	return false
}
func mustParseFile(r *http.Request, templateName string) *template.Template {
	prefix := "";
	if isMobile(r) {
		prefix = "m/"
	}
	tmpl, err := template.ParseFile("templates/" + prefix + templateName + ".html")
	if err != nil {
		fmt.Printf("error failed to parse template. %v\n", err)
	}
	return tmpl
}

func getUser(c appengine.Context, w http.ResponseWriter, r *http.Request) TwitterUser {
	sessionId := getSessionId(c, r)
	if sessionId == "" {
		c.Infof("no auth information. redirect to oauth confirmation.")
		return TwitterUser{}
	}
	user, err := FindUser(c, sessionId)
	if err != nil {
		c.Infof("OauthLikeHandler failed to find user. %v", err)
		return TwitterUser{}
	}
	return user
}

func getCookie(r *http.Request, name string) string {
	for _, c := range r.Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}
func getSessionId(c appengine.Context, r *http.Request) string {
	return getCookie(r, "id")
}

func getCommonMap(c appengine.Context, user TwitterUser) (map[string]interface{}, os.Error) {
	commonMap := make(map[string]interface{})

	commonMap["user"] = user
	commonMap["siteTitle"] = SiteTitle
	hashtagsForTicker, _ := GetPublicHashtags(c, map[string]interface{}{
		"length": 5,
	})
	commonMap["Common_HashtagsForTicker"] = hashtagsForTicker

	hashtagsHot, err := GetPublicHashtags(c, map[string]interface{}{
		"length": 5,
		"order": "-View",
	})
	if err != nil {
		c.Errorf("getCommonMap failed to retrieve hashtags: %v", err.String())
		return commonMap, err
	}
	commonMap["Common_HashtagsHot"] = hashtagsHot

	hashtagsNew, err := GetPublicHashtags(c, map[string]interface{}{
		"length": 5,
		"order": "-Date",
	})
	if err != nil {
		c.Errorf("getCommonMap failed to retrieve hashtags: %v", err.String())
		return commonMap, err
	}
	commonMap["Common_HashtagsNew"] = hashtagsNew
	
	return commonMap, nil
}

func FrontTop(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	user := getUser(c, w, r)
	resultMap, err := getCommonMap(c, user)
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

	var topTemplate = mustParseFile(r, "index")
	c.Debugf("parsedd")
	if err := topTemplate.Execute(w, resultMap); err != nil {
		c.Errorf("FrontTop failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}

func FrontSubject(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	path := r.URL.Path
	if len(path) < 4 {
		c.Errorf("FrontSubject 404: %v", path)
		ErrorPage(w, "お探しのページが見付かりませんでした。", http.StatusNotFound)
		return
	}
	hashtag := "#" + path[3:]

	user := getUser(c, w, r)
	resultMap, err := getCommonMap(c, user)
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
	sort := r.FormValue("sort")

	var tweets []Tweet
	tweets, err = GetTweetsByHashtag(c, hashtag, map[string]interface{}{
		"length": 20,
		"sort": sort,
	})
	if err != nil {
		c.Errorf("FrontSubject failed to retrieve tweets: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
	}
	hle := HashtagListElement{h, tweets}

	resultMap["sort"] = sort
	resultMap["elements"] = hle
	resultMap["encodedhashtag"] = Encode(hashtag[1:])
	var subjectTemplate = mustParseFile(r, "subject")
	if err := subjectTemplate.Execute(w, resultMap); err != nil {
		c.Errorf("FrontSubject failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
	c.Debugf("FrontSubject end")
}

func FrontHashtags(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	user := getUser(c, w, r)
	resultMap, err := getCommonMap(c, user)
	if err != nil {
		c.Errorf("FrontHashtags failed to retrieve resultMap: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}

	hashtags, err := GetHashtags(c, map[string]interface{}{
		"length": 300,
	})
	if err != nil {
		c.Errorf("FrontHashtags failed to search hashtags: %v", err)
		ErrorPage(w, "お探しのページが見付かりませんでした。", http.StatusNotFound)
		return
	}

	resultMap["hashtags"] = hashtags
	var hashtagsTemplate = mustParseFile(r, "hashtags")
	if err := hashtagsTemplate.Execute(w, resultMap); err != nil {
		c.Errorf("FrontHashtags failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}

func FrontHashtagsMore(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	pageStr := r.FormValue("page")
	page, _ := strconv.Atoi(pageStr)
	c.Debugf("FrontHashtags page: %d", page)

	hashtags, err := GetHashtags(c, map[string]interface{}{
		"page": page,
		"length": 300,
	})
	if err != nil {
		c.Errorf("FrontHashtags failed to search hashtags: %v", err)
		ErrorPage(w, "お探しのページが見付かりませんでした。", http.StatusNotFound)
		return
	}

	var hashtagsMoreTemplate = mustParseFile(r, "hashtags_more")
	if err := hashtagsMoreTemplate.Execute(w, map[string]interface{}{
				"page": page,
				"hashtags": hashtags,
			}); err != nil {
		c.Errorf("FrontHashtags failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}

func FrontAbout(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	user := getUser(c, w, r)
	resultMap, err := getCommonMap(c, user)
	if err != nil {
		c.Errorf("FrontHashtags failed to retrieve resultMap: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}

	var aboutTemplate = mustParseFile(r, "about")
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

	user := getUser(c, w, r)
	err := PointUpTweet(c, key, pointType, user)
	if err != nil {
		c.Errorf("PointUpHandler failed to point up. : %v", err)
		ErrorPage(w, "ポイントアップに失敗しました。", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, user.ScreenName)
}

func FavoriteHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("FavoriteHandler")

	if r.Method != "POST" {
		c.Errorf("FavoriteHandler method not supported.")
		ErrorPage(w, "ポイントアップに失敗しました。", http.StatusInternalServerError)
		return
	}
	key := r.FormValue("key")

	user := getUser(c, w, r)
	if err := PointUpTweet(c, key, "favorite", user); err != nil {
		c.Errorf("FavoriteHandler failed to point up. : %v", err)
		ErrorPage(w, "ポイントアップに失敗しました。", http.StatusInternalServerError)
		return
	}
	statusId := r.FormValue("statusId")
	if user.ScreenName == "" {
		c.Infof("FavoriteHandler failed to find user. empty user.")
		fmt.Fprint(w, "needs_oauth")
		return
	}
	if err := FavoriteTweetByUser(c, statusId, user); err != nil {
		c.Errorf("FavoriteHandler failed to point up. : %v", err)
		ErrorPage(w, "favoriteに失敗しました。", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, user.ScreenName)
}

func RetweetHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("RetweetHandler")

	if r.Method != "POST" {
		c.Errorf("RetweetHandler method not supported.")
		ErrorPage(w, "ポイントアップに失敗しました。", http.StatusInternalServerError)
		return
	}
	key := r.FormValue("key")

	user := getUser(c, w, r)
	if err := PointUpTweet(c, key, "retweet", user); err != nil {
		c.Errorf("RetweetHandler failed to point up. : %v", err)
		ErrorPage(w, "ポイントアップに失敗しました。", http.StatusInternalServerError)
		return
	}
	statusId := r.FormValue("statusId")
	if user.ScreenName == "" {
		c.Infof("RetweetHandler failed to find user. empty user.")
		fmt.Fprint(w, "needs_oauth")
		return
	}
	if err := RetweetTweetByUser(c, statusId, user); err != nil {
		c.Errorf("RetweetHandler failed to point up. : %v", err)
		ErrorPage(w, "retweetに失敗しました。", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, user.ScreenName)
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
	user := getUser(c, w, r)
	if user.ScreenName == "" {
		c.Infof("OauthLikeHandler failed to find user. empty user.")
		http.Redirect(w, r, url, 302)
		return
	}
	c.Debugf("OauthLikeHandler user: %v", user)

	err := LikeTweet(c, key, user)
	if err != nil {
		c.Errorf("OauthLikeHandler failed to point up. : %v", err)
		ErrorPage(w, "Likeに失敗しました。", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, 302)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("PostHandler")

	if r.Method != "POST" {
		c.Errorf("PostHandler method not supported.")
		ErrorPage(w, "Postに失敗しました。", http.StatusInternalServerError)
		return
	}

	// パラメータの保存 Oauthの後にリダイレクトするため。
	http.SetCookie(w, &http.Cookie{
		Name: "type",
		Value: "post",
		Path: "/",
	})
	status := r.FormValue("status")
	c.Debugf("PostHandler status: %s", status)
	http.SetCookie(w, &http.Cookie{
		Name: "status",
		Value: base64.StdEncoding.EncodeToString([]byte(status)),
		Path: "/",
	})
	hashtag := r.FormValue("hashtag")
	c.Debugf("PostHandler hashtag: %s", hashtag)
	http.SetCookie(w, &http.Cookie{
		Name: "hashtag",
		Value: base64.StdEncoding.EncodeToString([]byte(hashtag)),
		Path: "/",
	})
	url := r.FormValue("url")
	c.Debugf("PostHandler url: %s", url)
	http.SetCookie(w, &http.Cookie{
		Name: "url",
		Value: url,
		Path: "/",
	})
	user := getUser(c, w, r)
	if user.ScreenName == "" {
		c.Infof("PostHandler failed to find user. empty user.")
		fmt.Fprint(w, "needs_oauth")
		return
	}
	c.Debugf("PostHandler user: %v", user)

	tweet, err := PostTweetByUser(c, status, user)
	if err != nil {
		c.Errorf("PostHandler failed to point up. : %v", err)
		ErrorPage(w, "Postに失敗しました。", http.StatusInternalServerError)
		return
	}

	//新しいTweetの保存
	t := NewTweet(*tweet)
	t.Screen_name = tweet.User.Screen_name
	t.Profile_Image_Url = tweet.User.Profile_Image_Url
	t.UserId_Str = tweet.User.Id_Str
	SaveTweet(c, t, hashtag)
	var tweets = []Tweet{t}
	var subjectMoreTemplate = mustParseFile(r, "subject_more")
	if err := subjectMoreTemplate.Execute(w, map[string]interface{}{
				"Tweets": tweets,
			}); err != nil {
		c.Errorf("FrontSubjectMore failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}

func OauthPostHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("OauthPostHandler")

	hashtagBytes, _ := base64.StdEncoding.DecodeString(getCookie(r, "hashtag"))
	hashtag := string(hashtagBytes)
	statusBytes, _ := base64.StdEncoding.DecodeString(getCookie(r, "status"))
	status := string(statusBytes)
	url := getCookie(r, "url")
	if len(url) == 0 {
		url = "/"
	}
	c.Debugf("OauthPostHandler status: %s url: %s", status, url)
	user := getUser(c, w, r)
	if user.ScreenName == "" {
		c.Infof("OauthPostHandler failed to find user. empty user.")
		http.Redirect(w, r, url, 302)
		return
	}
	c.Debugf("OauthPostHandler user: %v", user)

	tweet, err := PostTweetByUser(c, status, user)
	if err != nil {
		c.Errorf("OauthPostHandler failed to point up. : %v", err)
		ErrorPage(w, "Postに失敗しました。", http.StatusInternalServerError)
		return
	}
	//新しいTweetの保存
	t := NewTweet(*tweet)
	t.Screen_name = tweet.User.Screen_name
	t.UserId_Str = tweet.User.Id_Str
	SaveTweet(c, t, hashtag)
	http.Redirect(w, r, url, 302)
}

func SignoutHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("SignoutHandler")

	url := getCookie(r, "url")
	if len(url) == 0 {
		url = "/"
	}
	c.Debugf("SignoutHandler url: %s", url)
	sessionId := getSessionId(c, r)
	if sessionId == "" {
		c.Infof("no auth information. redirect to oauth confirmation.")
	} else {
		if err := DeleteUser(c, sessionId); err != nil {
			c.Errorf("SignoutHandler failed to delete user: %v", err)
		}
		c.Debugf("SignoutHandler ok")
		http.SetCookie(w, &http.Cookie{
			Name: "id",
			Value: "",
			Path: "/",
		})
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
	http.SetCookie(w, &http.Cookie{
		Name: "type",
		Value: "like",
		Path: "/",
	})
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
	user := getUser(c, w, r)
	if user.ScreenName == "" {
		c.Infof("LikeHandler failed to find user. empty user.")
		fmt.Fprint(w, "needs_oauth")
		return
	}
	c.Debugf("LikeHandler user: %v", user)

	err := LikeTweet(c, key, user)
	if err != nil {
		c.Errorf("LikeHandler failed to point up. : %v", err)
		ErrorPage(w, "Likeに失敗しました。", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, user.ScreenName)
}

func FrontSubjectMore(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	hashtag := r.FormValue("hashtag")
	pageStr := r.FormValue("page")
	c.Debugf("FrontSubjectMore hashtag: %s", hashtag)
	page, _ := strconv.Atoi(pageStr)
	c.Debugf("FrontSubjectMore page: %d", page)
	sort := r.FormValue("sort")
	c.Debugf("FrontSubjectMore sort: %s", sort)

	var tweets []Tweet
	tweets, err := GetTweetsByHashtag(c, hashtag, map[string]interface{}{
		"length": 20,
		"page": page,
		"sort": sort,
	})
	if err != nil {
		c.Errorf("FrontSubjectMore failed to retrieve tweets: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
	}

	var subjectMoreTemplate = mustParseFile(r, "subject_more")
	if err := subjectMoreTemplate.Execute(w, map[string]interface{}{
				"Tweets":tweets,
			}); err != nil {
		c.Errorf("FrontSubjectMore failed to merge template: %v", err.String())
		ErrorPage(w, err.String(), http.StatusInternalServerError)
		return
	}
}
