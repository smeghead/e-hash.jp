package utamaru

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"http"
	"template"
	"time"
)

type Greeting struct {
	Author string
	Content string
	Date datastore.Time
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/sign", sign)
	http.HandleFunc("/cron/record_hashtags", RecordHashtags)
	http.HandleFunc("/cron/record_trends_hashtags", RecordTrendsHashtags)
	http.HandleFunc("/cron/crawle_hashtags", CrawleHashtags)
	http.HandleFunc("/worker/crawle_hashtag", WorkerCrawleHashtagHandler)
	http.HandleFunc("/home_test", HomeTestHandler)
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Hashtag").Order("-Count").Limit(10)
	hashtags := make([]Hashtag, 0, 10)
	if _, err := q.GetAll(c, &hashtags); err != nil {
		c.Errorf("root failed to crawle by hashtag: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	if err := guestbookTemplate.Execute(w, hashtags); err != nil {
		c.Errorf("root failed to crawle by hashtag: %v", err.String())
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}

var guestbookTemplate = template.MustParse(guestbookTemplateHtml, nil)

var guestbookTemplateHtml = `
<html>
  <body>
    <h1>hashtags</h1>
    {.repeated section @}
      <p><b>{Name|html}</b> {Count|html}</p>
    {.end}
    <form action="/worker/crawle_hashtag" method="POST">
      <input type="text" name="hashtag" value="#twitter">
      <input type="submit" value="crawle_hashtag">
    </form>

  </body>
</html>
`

func sign(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := Greeting {
		Content: r.FormValue("content"),
		Date: datastore.SecondsToTime(time.Seconds()),
	}
	if u := user.Current(c); u != nil {
		g.Author = u.String()
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey("Greeting"), &g)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

var signTemplate = template.MustParse(signTemplateHtml, nil)

var signTemplateHtml = `
<html>
  <body>
    <p>You wrote:</p>
    <pre>{@|html}</pre>
  </body>
</html>
`

func HomeTestHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("HomeTestHandler")
	err := HomeTest(c)
	if err != nil {
		c.Errorf("HomeTestHandler failed to post: %v", err.String())
	}
	fmt.Fprint(w, "end");
}
