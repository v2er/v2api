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

// Profile 个人资料
type Profile struct {
	UserName      string
	UserUrl       string
	Avatar        string
	FavNodes      int
	FavTopics     int
	Following     int
	Notifications int
	Balance       *Balance
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
	Index    int
	UserName string
	Balance  *Balance
}

// Balance
type Balance struct {
	Gold   int
	Silver int
	Bronze int
	Money  float32
}

// Content
type Content struct {
	Topic        *Topic
	Body         string
	Clicks       int
	Favorites    int
	Thanks       int
	Replies      []*Reply
	ReplyTotal   int
	ReplyTime    time.Time
	ReplyPage    int
	ReplyPageMax int

	// TODO: 附言
}

// Reply
type Reply struct {
	Author      string
	AuthorUrl   string
	Avatar      string
	Number      int
	Content     string
	Publish     string
	PublishTime time.Time
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

	D, _ := regNum(`(\d+)天`, publish)
	H, _ := regNum(`(\d+)小时`, publish)
	M, _ := regNum(`(\d+)分钟`, publish)

	dur := time.Duration(D)*time.Hour*24 +
		time.Duration(H)*time.Hour +
		time.Duration(M)*time.Minute

	return time.Now().Add(-dur), nil
}

func regNum(reg, str string) (int, error) {
	res := regexp.MustCompile(reg).FindStringSubmatch(str)
	if len(res) > 0 {
		return strconv.Atoi(res[1])
	}
	return 0, nil
}

func parseBalance(s *goquery.Selection) (*Balance, error) {
	nums := regexp.MustCompile(`\d+`).FindAllString(s.Text(), -1)

	imgs := s.Find("img")
	if imgs.Length() == 0 {
		// 兼容[消费排行]格式
		money := s.Text()
		removeSpace(&money)
		if strings.HasPrefix(money, "$") {
			money = strings.TrimLeft(money, "$")
			value, _ := strconv.ParseFloat(money, 32)
			return &Balance{Money: float32(value)}, nil
		}
	}
	if imgs.Length() != len(nums) {
		return nil, errors.New("Parse balance failed")
	}

	i := 0
	b := &Balance{}

	imgs.Each(func(_ int, img *goquery.Selection) {
		src, _ := img.Attr("src")
		num, _ := strconv.Atoi(nums[i])
		i++
		switch {
		case strings.Contains(src, "gold"):
			b.Gold = num
		case strings.Contains(src, "silver"):
			b.Silver = num
		case strings.Contains(src, "bronze"):
			b.Bronze = num
		}
	})

	return b, nil
}

func parseTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-0215:04:05+08:00", s, time.Local)
}

func onError(err error) {
	if err != nil {
		panic(err)
	}
}
