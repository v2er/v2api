package v2api

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

const (
	URL_HOME   = "https://www.v2ex.com"
	URL_RECENT = URL_HOME + "/recent"
	URL_PLANES = URL_HOME + "/planes"
)

var (
	ErrNotLogin = errors.New("Not login")
)

var DefaultClient *Client

type Topic struct {
	Title       string
	Link        string
	Author      string
	AuthorUrl   string
	Avatar      string
	Votes       int
	Comments    int
	Reply       string
	Node        string
	NodeUrl     string
	Publish     string
	PublishTime time.Time
}

// Node 节点
// https://www.v2ex.com/planes
type Node struct {
	Name   string
	URL    string
	Type   string
	TypeCN string
}

// Community 社区数据
type Community struct {
	// 首页右侧
	Members  int // 会员
	Topics   int // 主题
	Comments int // 回复

	// 首页底部
	Version   string // 版本
	Online    int    // 当前在线
	OnlineMax int    // 最高在线
}

// Leaderboard 排行榜
// 财富排行榜 https://www.v2ex.com/top/rich
// 消费排行榜 https://www.v2ex.com/top/player
type Leaderboard struct {
}

func init() {
	DefaultClient = &Client{}
}

func parseSelection(s *goquery.Selection) (*Topic, error) {
	t := &Topic{}

	t.Title = s.Find(".topic-link").Text()

	t.Link, _ = s.Find(".topic-link").Attr("href")
	completeURL(&t.Link)

	t.Author = s.Find("strong a").Text()

	t.AuthorUrl, _ = s.Find("strong a").Attr("href")
	completeURL(&t.AuthorUrl)

	t.Avatar, _ = s.Find(".avatar").Attr("src")
	if len(t.Avatar) > 0 {
		t.Avatar = "https:" + t.Avatar
	}

	votes := s.Find(".votes").Text()
	removeSpace(&votes)
	t.Votes, _ = strconv.Atoi(votes)

	count := s.Find(".count_livid").Text()
	t.Comments, _ = strconv.Atoi(count)

	t.Node = s.Find(".node").Text()

	t.NodeUrl, _ = s.Find(".node").Attr("href")
	completeURL(&t.NodeUrl)

	infoText := s.Find(".topic_info").Text()
	removeSpace(&infoText)
	infoSlice := strings.Split(infoText, "•")

	if len(infoSlice) > 2 {
		t.Publish = infoSlice[2]
		t.PublishTime, _ = publishToTime(t.Publish)
	}

	if len(infoSlice) > 3 {
		t.Reply = strings.TrimLeft(infoSlice[3], "最后回复来自")
	}

	return t, nil
}

func completeURL(s *string) {
	if len(*s) > 0 {
		*s = URL_HOME + *s
	}
}

func removeSpace(s *string) {
	v := *s
	v = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, v)
	v = strings.ReplaceAll(v, " ", "")
	*s = v
}

// TODO:
func publishToTime(publish string) (t time.Time, err error) {
	return
}

func onError(err error) {
	if err != nil {
		panic(err)
	}
}
