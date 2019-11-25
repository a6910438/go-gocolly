package spider

import (
	"bytes"
	"encoding/json"
	"github.com/a6910438/go-logger"
	"github.com/bitly/go-simplejson"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
	"io/ioutil"
	"net"
	"net/http"
	"spider/types"
	"strconv"
	"strings"
	"time"
)

type dbBaser interface {
	Add(info types.Information) (id uint64, err error)
	GetInfoByTitleOrContext(title, context string) (*types.Information, error)
	UpdateInfoById(information *types.Information) (err error)
}

type Spider struct {
	db   dbBaser
	site string
	host string
}

func NewSpider(db dbBaser, site, host string) (*Spider, error) {
	return &Spider{
		db:   db,
		site: site,
		host: host,
	}, nil
}

/**
	构建爬虫请求并返回一个页面收集对象
 */
func (s *Spider) NewClient() *colly.Collector {
	c := colly.NewCollector()
	// 设置HTTP常规设置
	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // 超时时间
			KeepAlive: 30 * time.Second, // keepAlive 超时时间
			Deadline:  time.Now().Add(30 * time.Second),
		}).DialContext,
		MaxIdleConns:          100,              // 最大空闲连接数
		IdleConnTimeout:       90 * time.Second, // 空闲连接超时
		TLSHandshakeTimeout:   10 * time.Second, // TLS 握手超时
		ExpectContinueTimeout: 1 * time.Second,
	})

	c.OnError(func(_ *colly.Response, e error) {
		// 创建爬虫请求时出错,停留五分钟再请求
		logger.Errorf("New Client went wrong: %v, site url is: ", e, s.site)
		time.Sleep(time.Second * 300)
	})

	// 设置请求头
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "*/*")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Host", s.host)
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.97 Safari/537.36")
	})
	return c
}

/**
	创建一个Http请求,Tokenview的数据无法通过html拿取就利用http接口获取
 */
func (s *Spider) NewHTTPClient() ([]byte, error) {
	// 创建一个GET请求
	now := time.Now().Unix()
	site := strings.Replace(s.site, "{{time}}", strconv.FormatInt(now, 10), -1)
	resp, err := http.Get(site)
	if err != nil {
		return nil, errors.WithMessage(err, "create http request exception")
	}

	defer resp.Body.Close()
	// 读取返回的数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "read request inputstream exception")
	}

	return body, nil
}

/**
	抓取币世界的资讯数据
 */
func (s *Spider) SpiderBishijie(ch chan types.Information) {
	for {
		// 构建新请求
		collector := s.NewClient()
		// 根据HTML里数据的位置,抓取对应的内容
		collector.OnHTML(".content", func(e *colly.HTMLElement) {
			context := e.ChildText(".h63")
			time := e.ChildText("span")
			title := strings.Replace(e.ChildText("h3"), " ", "", -1)
			title = strings.Replace(title, "\n", "", -1)
			title = strings.Replace(title, time, "", -1)
			//判断是否为空字符,空字符不处理
			if title != "" && context != "" {
				s.sendInfo(ch, types.Author, title, context, types.InformationChinese)
			}
		})
		collector.Visit(s.site)
		time.Sleep(time.Second * 30)
	}
}

/**
	抓取巴比特的资讯数据
 */
func (s *Spider) Spider8btc(ch chan types.Information) {
	for {
		// 构建新请求
		collector := s.NewClient()
		// 根据HTML里数据的位置,抓取对应的内容
		collector.OnHTML(".flash-wrap", func(e *colly.HTMLElement) {
			title := e.ChildText(".flash-item__title")
			context := e.ChildText(".flash-item__content")
			//判断是否为空字符,空字符不处理
			if title != "" && context != "" {
				titles := strings.Split(title, "】")
				title = strings.Replace(titles[0], "【", "", -1)
				s.sendInfo(ch, types.Author, title, context, types.InformationChinese)
			}
		})
		collector.Visit(s.site)
		time.Sleep(time.Second * 10)
	}
}

/**
	抓取Tokenview资讯数据
 */
func (s *Spider) SpiderTokenview(ch chan types.Information) {
	for {
		data, err := s.NewHTTPClient()
		if err != nil {
			//发生异常先休息一分钟,再跳回重试
			logger.Error("new http request", "err", err)
			time.Sleep(time.Second * 60)
			continue
		}
		//判断语言类型,如果URL中带有中文标识就标记为中文资讯
		//否则就走英文
		index := strings.Index(s.site, "cn")
		if index == -1 {
			err = s.startTokenView(data, ch, types.InformationEnglish)
		} else {
			err = s.startTokenView(data, ch, types.InformationChinese)
		}
		if err != nil {
			//发生异常先休息一分钟,再跳回重试
			logger.Error("start spider tokenview", "err", err)
			time.Sleep(time.Second * 60)
			continue
		}
		time.Sleep(time.Second * 100)
	}
}

/**
	请求获取的json数据转换成资讯结构体
 */
func (s *Spider) startTokenView(dataJ []byte, ch chan types.Information, lang types.InfomationLanguage) error {
	res, err := simplejson.NewJson(dataJ)
	if err != nil {
		return errors.WithMessage(err, "tokenview new simple json exception")
	}
	//获取标题内容数据
	temps, err := res.Get("templates").Array()
	if err != nil {
		return errors.WithMessage(err, "tokenview parse json exception")
	}
	// 获取数据模板
	items, err := res.Get("items").Array()
	for ti := range temps {
		tempID, _ := res.Get("templates").GetIndex(ti).Get("id").Int()
		// 取"0"是因为json值里面 "0"是真正的标题和内容 "1"是转义符无需管理
		contents, _ := res.Get("templates").GetIndex(ti).Get("content").Get("0").String()
		cs := strings.Split(contents, "\n")
		// 解析失败就解析下一条资讯
		if err != nil {
			continue
		}
		// 开始解析并发送消息
		s.startParseJSONAndSendInfo(ch, res, items, cs[0], cs[1], ti, tempID, lang)
	}
	return nil
}

/**
	对请求得到来的json进行解析
	ch  通道
	res json对象
	items json对象的子集合
	title 标题
	context 未处理的正文
	index 格式化的数据所在json的下标
	tempID  模板ID
 */
func (s *Spider) startParseJSONAndSendInfo(ch chan types.Information, res *simplejson.Json, items []interface{}, title, context string, index, tempID int, lang types.InfomationLanguage) {
	for ii := range items {
		tID, _ := res.Get("items").GetIndex(ii).Get("templateId").Int()
		// 两个模板必须是一样才认定属于同一条资讯
		if tID == tempID {
			m, _ := res.Get("items").GetIndex(ii).Get("data").Map()
			for k := range m {
				// 由于JSON结构复杂,得走两层判断逻辑
				values, _ := res.Get("items").GetIndex(ii).Get("data").Get(k).Map()
				if values != nil {
					// 转哈希结构得逻辑处理
					context, title = mapTypeToString(values, title, context)
				} else {
					// 转数组结构的逻辑处理
					context, title = arrayTypeToString(res, ii, index, k, title, context)
				}
			}
			// 英文要处理一下,返回的数据里面有两个字是中文的,要统统换成英文
			if lang == types.InformationEnglish {
				context, title = specialEnglishContext(context, title)
			}
			// 处理完成之后通知主线程处理该资讯
			s.sendInfo(ch, types.Author, title, context, lang)
			break
		}
	}
}

/**
	Tokenview爬取的json里面 由于结构复杂,只能判断其类型做相对应的解析
	其中有两种(1.数组里面放结构体, 2.无数组的结构体)
	有数组的结构体进行解析
	context 未处理的正文
 */
func arrayTypeToString(res *simplejson.Json, pindex, sindex int, key, title, context string) (string, string) {
	var buffer bytes.Buffer
	var appendFormatStr string
	values, _ := res.Get("items").GetIndex(pindex).Get("data").Get(key).Array()
	buffer.WriteString(context + `\n`)
	for vm := range values {
		vs, _ := res.Get("items").GetIndex(pindex).Get("data").Get(key).GetIndex(vm).Map()
		// 判断有没有要循环追加的内容，如果有 就在循环items的时候把值递进去, "1"取的是要替换的数据内容 固定不变与上面的0有两层意思
		append := res.Get("templates").GetIndex(sindex).Get("content").Get("1")
		if append != nil {
			appendFormatStr, _ = res.Get("templates").GetIndex(sindex).Get("content").Get("1").String()
		}
		for v, s := range vs {
			// 把key为类似这种{{btc}}替换成对应的value
			value := jsonTypeToString(s)
			old := "{{" + v + "}}"
			if appendFormatStr != "" {
				appendFormatStr = strings.Replace(appendFormatStr, old, value, -1)
			}
			title = strings.Replace(title, old, value, -1)
		}
		if appendFormatStr != "" {
			buffer.WriteString(appendFormatStr + `\n`)
		}
	}
	return buffer.String(), title
}

/**
	无数组的结构体进行解析
 */
func mapTypeToString(m map[string]interface{}, title, context string) (string, string) {
	for b, v := range m {
		now := jsonTypeToString(v)
		old := "{{" + b + "}}"
		context = strings.Replace(context, old, now, -1)
		title = strings.Replace(title, old, now, -1)
	}
	return context, title
}

/**
	No.二 is BTC,with $ 286.69 hundred million in 24-hour tunrover,which accounts for 35.17% of the total.
	下降0.03%;BCH is $223.26,下降8.72%;
	部分英文数据带有中文内容需手动处理
 */
func specialEnglishContext(context, title string) (string, string) {
	context = strings.Replace(context, "一", "1 ", -1)
	context = strings.Replace(context, "二", "2 ", -1)
	context = strings.Replace(context, "三", "3 ", -1)
	context = strings.Replace(context, "四", "4 ", -1)
	context = strings.Replace(context, "五", "5 ", -1)
	context = strings.Replace(context, "下降", "down ", -1)
	context = strings.Replace(context, "上升", "up ", -1)
	title = strings.Replace(title, "下降", "down ", -1)
	title = strings.Replace(title, "上升", "up ", -1)
	return context, title
}

/**
    json类型的数字转换成字符串
	v json类型
 */
func jsonTypeToString(v interface{}) string {
	var now string
	switch v.(type) {
	case json.Number:
		now = v.(json.Number).String()
	case string:
		now = v.(string)
	default:
		// TODO Something
	}
	return now
}

/**
	资讯存放到数据库中,去重复标题和内容的资讯
 */
func (s *Spider) AddInfodata(info types.Information) (id uint64, err error) {
	i, err := s.db.GetInfoByTitleOrContext(info.Title, info.Context)
	// 通过资讯标题发现有无重复数据
	if i.Id != 0 {
		//指定ID更新数据
		info.Id = i.Id
		err = s.db.UpdateInfoById(&info)
		if err != nil {
			return i.Id, err
		}
		return i.Id, nil
	}
	// 直接录入数据库
	return s.db.Add(info)
}

/**
	封装资讯信息通知主线程即时处理
 */
func (s *Spider) sendInfo(ch chan types.Information, author, title, context string, lang types.InfomationLanguage) {
	// 处理完成之后通知主线程处理该资讯
	ch <- types.Information{
		Author:     author,
		Title:      title,
		Context:    context,
		CreateTime: time.Now().Unix(),
		Lang:       lang,
	}
}
