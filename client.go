package v2api

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Client struct {
	cookie string
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) SetCookie(cookie string) {
	c.cookie = cookie
}

// Latest 首页最新
func (c *Client) Latest() (topics []*Topic, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	doc, err := c.queryDocument(URL_HOME)
	onError(err)

	topics = make([]*Topic, 0)
	doc.Find(".cell.item").Each(func(i int, s *goquery.Selection) {
		topic, err := parseSelection(s)
		onError(err)
		topics = append(topics, topic)
	})

	return
}

// Recent 最近主题
func (c *Client) Recent() {}

func (c *Client) queryDocument(url string) (doc *goquery.Document, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Add("cookie", c.cookie)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	return goquery.NewDocumentFromResponse(res)
}
