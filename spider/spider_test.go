package spider

import (
	"github.com/stretchr/testify/require"
	"spider/cmd/app/config"
	"spider/database/mysql"
	"spider/types"
	"testing"
)

func TestSpider(t *testing.T) {
	mysql, err := mysql.NewMysql(config.DbConfig{
		User:     "test",
		Password: "123456",
		Host:     "192.168.0.210",
		Port:     3306,
		Dbname:   "news",
	})
	require.Nil(t, err)
	spider, err := NewSpider(mysql, "https://www.bishijie.com/kuaixun/", "https://www.bishijie.com/")
	require.Nil(t, err)
	entokenview, err := NewSpider(mysql, "https://tokenview.com/v2api/news/list/all/{{time}}/en", "https://tokenview.com/")
	require.Nil(t, err)
	cntokenview, err := NewSpider(mysql, "https://tokenview.com/v2api/news/list/all/{{time}}/cn", "https://tokenview.com/")
	require.Nil(t, err)
	btc, err := NewSpider(mysql, "https://www.8btc.com/flash", "https://www.8btc.com/")
	require.Nil(t, err)
	ch := make(chan types.Information)
	defer close(ch)
	go spider.SpiderBishijie(ch)
	go entokenview.SpiderTokenview(ch)
	go cntokenview.SpiderTokenview(ch)
	go btc.Spider8btc(ch)
	for {
		select {
		case data := <-ch:
			spider.AddInfodata(data)
		}
	}
}
