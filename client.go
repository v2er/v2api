package v2api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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

func (c *Client) HasLoggedIn() bool {
	return len(c.cookie) > 0
}

func (c *Client) mustLogin() {
	if !c.HasLoggedIn() {
		onError(ErrNotLogin)
	}
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
		topic, err := parseTopic(s)
		onError(err)
		topics = append(topics, topic)
	})

	return
}

// Hots 今日热议主题
func (c *Client) Hots() (topics []*Topic, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	doc, err := c.queryDocument(URL_HOME)
	onError(err)

	topics = make([]*Topic, 0)
	doc.Find("#TopicsHot table").Each(func(i int, s *goquery.Selection) {
		t := &Topic{}

		a := s.Find(".item_hot_topic_title a")
		t.Title = a.Text()
		t.Link, _ = a.Attr("href")
		completeURL(&t.Link)

		t.AuthorUrl, _ = s.Find("a").Eq(0).Attr("href")
		t.Author = strings.TrimLeft(t.AuthorUrl, "/member/")
		completeURL(&t.AuthorUrl)
		t.Avatar, _ = s.Find(".avatar").Attr("src")
		completeURL(&t.Avatar)

		topics = append(topics, t)
	})

	return
}

// Member 会员数据
func (c *Client) Member(name string) (mem *Member, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	doc, err := c.queryDocument(URL_MEMBER + name)
	onError(err)

	slt := doc.Find("#Main .box").Eq(0)

	str := slt.Find(".gray").Text()
	removeSpace(&str)

	reg := regexp.MustCompile(`V2EX第(\d+)号会员，加入于(.+)今日活跃度排名(\d+)`)
	res := reg.FindStringSubmatch(str)
	number, join, rank := res[1], res[2], res[3]

	mem = &Member{}
	mem.Name = name
	mem.Bio = slt.Find(".bigger").Text()
	mem.Avatar, _ = slt.Find(".avatar").Attr("src")
	completeURL(&mem.Avatar)
	mem.Number, _ = strconv.Atoi(number)
	mem.Join = join
	mem.JoinTime, _ = parseTime(join)
	mem.Rank, _ = strconv.Atoi(rank)
	mem.Online = slt.Find(".online").Length() > 0

	return
}

// Profile 个人资料
func (c *Client) Profile() (pro *Profile, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	c.mustLogin()

	doc, err := c.queryDocument(URL_HOME)
	onError(err)

	slt := doc.Find("#Rightbar .box").Eq(0)

	tmp := slt.Find("a .bigger")
	favNodes, favTopics, following := tmp.Eq(0).Text(), tmp.Eq(0).Text(), tmp.Eq(0).Text()

	notifications := slt.Find(".inner a.fade").Text()
	notifications = strings.TrimRight(notifications, " 条未读提醒")

	pro = &Profile{}
	pro.UserName = slt.Find(".bigger a").Text()
	pro.UserUrl = "/member/" + pro.UserName
	completeURL(&pro.UserUrl)
	pro.Avatar, _ = slt.Find("img.avatar").Attr("src")
	completeURL(&pro.Avatar)
	pro.FavNodes, _ = strconv.Atoi(favNodes)
	pro.FavTopics, _ = strconv.Atoi(favTopics)
	pro.Following, _ = strconv.Atoi(following)
	pro.Notifications, _ = strconv.Atoi(notifications)
	pro.Balance, _ = parseBalance(slt.Find("a.balance_area"))

	return
}

// Community 社区数据
func (c *Client) Community() (com *Community, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	doc, err := c.queryDocument(URL_HOME)
	onError(err)

	dom := doc.Find("#Rightbar div.box").Last().Find("strong")
	members, topics, comments := dom.Eq(0).Text(), dom.Eq(1).Text(), dom.Eq(2).Text()

	str := doc.Find("#Bottom").Text()
	removeSpace(&str)
	reg := regexp.MustCompile(`(\d+)人在线最高记录(\d+).+VERSION:(.+?)·`)
	res := reg.FindStringSubmatch(str)
	online, onlineMax, version := res[1], res[2], res[3]

	com = &Community{}
	com.Members, _ = strconv.Atoi(members)
	com.Topics, _ = strconv.Atoi(topics)
	com.Comments, _ = strconv.Atoi(comments)
	com.Online, _ = strconv.Atoi(online)
	com.OnlineMax, _ = strconv.Atoi(onlineMax)
	com.Version = version

	return
}

// TopRich 财富排行
func (c *Client) TopRich() ([]*Leaderboard, error) {
	return c.getLeaderboardList(URL_HOME + "/top/rich")
}

// TopPlay 消费排行
func (c *Client) TopPlay() ([]*Leaderboard, error) {
	return c.getLeaderboardList(URL_HOME + "/top/player")
}

func (c *Client) getLeaderboardList(url string) (lbs []*Leaderboard, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	doc, err := c.queryDocument(url)
	onError(err)

	lbs = make([]*Leaderboard, 0)

	doc.Find("#Main .box .inner tr").Each(func(i int, s *goquery.Selection) {
		if i%2 > 0 || i > 48 {
			return
		}

		str := s.Find("h2").Text()
		removeSpace(&str)
		arr := strings.Split(str, ".")

		lb := &Leaderboard{}
		lb.Index, _ = strconv.Atoi(arr[0])
		lb.UserName = arr[1]
		lb.Balance, _ = parseBalance(s.Find(".balance_area"))

		lbs = append(lbs, lb)
	})

	return
}

// Planes 位面列表
func (c *Client) Planes() (nodes []*Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	doc, err := c.queryDocument(URL_PLANES)
	onError(err)

	nodes = make([]*Node, 0)
	doc.Find("#Main .box").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		str := s.Find(".header").Text()
		removeSpace(&str)
		str = strings.Split(str, "•")[0]

		fr := s.Find(".fr").Text()
		removeSpace(&fr)
		fr = strings.Split(fr, "•")[0]

		cn := strings.TrimRight(str, fr)

		s.Find(".inner a").Each(func(_ int, a *goquery.Selection) {
			node := &Node{}
			node.Name = a.Text()
			node.URL, _ = a.Attr("href")
			completeURL(&node.URL)
			node.Type = fr
			node.TypeCN = cn
			nodes = append(nodes, node)
		})
	})

	return
}

// Recent 最近主题
func (c *Client) Recent(page int) (list *List, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	if !c.HasLoggedIn() {
		return nil, ErrNotLogin
	}

	url := URL_RECENT
	if page > 0 {
		url += "?p=" + strconv.Itoa(page)
	}

	doc, err := c.queryDocument(url)
	onError(err)

	topics := make([]*Topic, 0)
	doc.Find(".cell.item").Each(func(i int, s *goquery.Selection) {
		topic, err := parseTopic(s)
		onError(err)
		topics = append(topics, topic)
	})

	total := doc.Find("#Main .fade").Text()
	total = regexp.MustCompile(`\d+`).FindString(total)

	pageMax := doc.Find(".page_normal").Last().Text()
	removeSpace(&pageMax)

	list = &List{}
	list.Topics = topics
	list.Total, _ = strconv.Atoi(total)
	list.Page = page
	list.PageMax, _ = strconv.Atoi(pageMax)

	return
}

// Node 节点主题
func (c *Client) Node(node string, page int) (list *List, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	if !c.HasLoggedIn() {
		return nil, ErrNotLogin
	}
	if len(node) == 0 {
		return nil, ErrNodeNotExist
	}

	url := URL_NODE + node
	if page > 0 {
		url += "?p=" + strconv.Itoa(page)
	}

	doc, err := c.queryDocument(url)
	onError(err)

	topics := make([]*Topic, 0)
	doc.Find("#TopicsNode .cell").Each(func(i int, s *goquery.Selection) {
		topic, err := parseTopic(s)
		onError(err)
		topics = append(topics, topic)
	})

	total := doc.Find(".fr.f12 strong").Text()

	pageMax := doc.Find(".page_normal").Last().Text()
	removeSpace(&pageMax)

	list = &List{}
	list.Topics = topics
	list.Total, _ = strconv.Atoi(total)
	list.Page = page
	list.PageMax, _ = strconv.Atoi(pageMax)
	list.NodeBio = doc.Find(".node_info span.f12").Text()
	list.NodeImg, _ = doc.Find(".node_avatar img").Attr("src")
	completeURL(&list.NodeImg)

	return
}

// Content
// replyPage 回复列表页数 从1开始 为0则最新页
func (c *Client) Content(id int64, replyPage int) (ctt *Content, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	ctt = &Content{}

	url := fmt.Sprintf("%s/t/%d", URL_HOME, id)
	if replyPage > 0 {
		url += "?p=" + strconv.Itoa(replyPage)
	}

	doc, err := c.queryDocument(url)
	onError(err)

	slt := doc.Find("#Main")

	ctt.Body, _ = slt.Find(".topic_content").Html()

	sltPage := slt.Find("a.page_current").Eq(0)
	if sltPage.Length() == 0 {
		ctt.ReplyPage = 1
		ctt.ReplyPageMax = 1
	} else {
		ctt.ReplyPage = sltPage.Index() + 1
		ctt.ReplyPageMax = sltPage.Parent().ChildrenFiltered("a").Length()
	}

	stats := slt.Find(".topic_stats").Text()
	removeSpace(&stats)
	ctt.Clicks, _ = regNum(`(\d+)次点击`, stats)
	ctt.Thanks, _ = regNum(`(\d+)人感谢`, stats)
	ctt.Favorites, _ = regNum(`(\d+)人收藏`, stats)

	topic := &Topic{}
	topic.Title = slt.Find(".header h1").Text()
	topic.Link = url
	topic.Author = slt.Find(".header .gray a").Text()
	topic.AuthorUrl, _ = slt.Find(".header .gray a").Attr("href")
	completeURL(&topic.AuthorUrl)
	topic.Avatar, _ = slt.Find("img.avatar").Attr("src")
	completeURL(&topic.Avatar)

	votes := slt.Find(".votes").Text()
	removeSpace(&votes)
	topic.Votes, _ = strconv.Atoi(votes)

	a_ := slt.Find(".chevron").Next()
	if a_.Is("a") {
		topic.Node = a_.Text()
		topic.NodeUrl, _ = a_.Attr("href")
		completeURL(&topic.NodeUrl)
	}

	tmp := slt.Find(".header .gray").Text()
	removeSpace(&tmp)
	arr := strings.Split(tmp, "·")
	if len(arr) > 1 {
		topic.Publish = arr[1]
		topic.PublishTime, _ = publishToTime(arr[1])
	}

	ctt.Topic = topic

	slt.Find(".subtle").Each(func(_ int, s *goquery.Selection) {
		str := s.Find(".fade").Text()
		removeSpace(&str)

		ps := &Postscript{}
		ps.Content, _ = s.Find(".topic_content").Html()
		ps.Publish = strings.Split(str, "·")[1]
		ps.PublishTime, _ = publishToTime(ps.Publish)

		ctt.Postscripts = append(ctt.Postscripts, ps)
	})

	noReply := slt.Find(".box.transparent").Length() > 0
	if noReply {
		return
	}

	box := slt.Find(".box").Eq(1)

	str := box.Find(".cell .gray").First().Text()
	removeSpace(&str)
	reg := regexp.MustCompile(`(\d+)回复\|直到(.+)`)
	res := reg.FindStringSubmatch(str)
	if len(res) == 3 {
		ctt.ReplyTotal, _ = strconv.Atoi(res[1])
		ctt.ReplyTime, _ = parseTime(res[2])
	}

	// reply list
	box.Find(".cell").Each(func(_ int, s *goquery.Selection) {
		if id, _ := s.Attr("id"); !strings.HasPrefix(id, "r_") {
			return
		}

		reply := &Reply{}

		reply.Author = s.Find("a.dark").Text()
		reply.AuthorUrl, _ = s.Find("a.dark").Attr("href")
		completeURL(&reply.AuthorUrl)
		reply.Avatar, _ = s.Find("img.avatar").Attr("src")
		completeURL(&reply.Avatar)
		reply.Number, _ = strconv.Atoi(s.Find(".no").Text())
		reply.Content, _ = s.Find(".reply_content").Html()

		ago := s.Find(".ago").Text()
		removeSpace(&ago)
		if si := strings.Index(ago, "via"); si > 0 {
			reply.Publish = ago[:si]
		} else {
			reply.Publish = ago
		}
		reply.PublishTime, _ = publishToTime(reply.Publish)

		ctt.Replies = append(ctt.Replies, reply)
	})

	return
}

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
