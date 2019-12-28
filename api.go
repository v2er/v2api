package v2api

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

const (
	URL_HOME   = "https://www.v2ex.com"
	URL_NODE   = URL_HOME + "/go/"
	URL_RECENT = URL_HOME + "/recent"
	URL_PLANES = URL_HOME + "/planes"
	URL_MEMBER = URL_HOME + "/member/"
)

var (
	ErrNotLogin = errors.New("Not login")

	ErrNodeNotExist = errors.New("Node is not exist")
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

type List struct {
	Topics  []*Topic
	Total   int
	Page    int
	PageMax int
	NodeBio string
	NodeImg string
}

// Node 节点
// https://www.v2ex.com/planes
type Node struct {
	Name   string
	URL    string
	Type   string
	TypeCN string
}

// Member 会员数据
// https://www.v2ex.com/member/livid
type Member struct {
	Name     string
	Bio      string
	Avatar   string
	Number   int
	Join     string
	JoinTime time.Time
	Rank     int
	Online   bool
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
	completeURL(&t.Avatar)

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

	replyPrefix := "最后回复来自"
	replyIndex := 0
	for i, v := range infoSlice {
		if strings.HasPrefix(v, replyPrefix) {
			replyIndex = i
		}
	}

	if replyIndex > 0 {
		t.Reply = strings.TrimLeft(infoSlice[replyIndex], replyPrefix)
		t.Publish = infoSlice[replyIndex-1]
	} else {
		t.Publish = infoSlice[len(infoSlice)-1]
	}
	t.PublishTime, _ = publishToTime(t.Publish)

	return t, nil
}

func completeURL(s *string) {
	if strings.HasPrefix(*s, "//") {
		*s = "https:" + *s
	} else if strings.HasPrefix(*s, "/") {
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

func publishToTime(publish string) (t time.Time, err error) {
	if !strings.HasSuffix(publish, "前") {
		return t, errors.New("Can't to parse time: " + publish)
	}

	getNum := func(s string) int {
		res := regexp.MustCompile(s).FindStringSubmatch(publish)
		if len(res) > 0 {
			num, _ := strconv.Atoi(res[1])
			return num
		}
		return 0
	}

	D := getNum(`(\d+)天`)
	H := getNum(`(\d+)小时`)
	M := getNum(`(\d+)分钟`)

	dur := time.Duration(D)*time.Hour*24 +
		time.Duration(H)*time.Hour +
		time.Duration(M)*time.Minute

	return time.Now().Add(-dur), nil
}

func onError(err error) {
	if err != nil {
		panic(err)
	}
}
