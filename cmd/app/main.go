package main

import (
	"flag"
	"fmt"
	"github.com/a6910438/go-logger"
	"github.com/judwhite/go-svc/svc"
	"spider/cmd/app/config"
	"spider/database/mysql"
	"spider/log"
	"spider/spider"
	"spider/types"
)

type program struct {
	bWorld      *spider.Spider // 币世界
	tokenViewEn *spider.Spider //TokenView英文
	tokenViewCn *spider.Spider //TokenView中文
	bBtc        *spider.Spider // 巴比特
}

func (p *program) Init(env svc.Environment) error {
	//初始化日志
	if err := log.Init(config.Cfg.Log.Path, config.Cfg.Log.File, config.Cfg.Log.Level); err != nil {
		fmt.Println("init log error:", err)
		return err
	}
	//初始化数据库
	mysql, err := mysql.NewMysql(config.Cfg.Db)
	if err != nil {
		logger.Errorf("init db error: %v", err)
		return err
	}

	//初始化所有爬虫程序
	bWorld, err := spider.NewSpider(mysql, config.Cfg.BiWorldUrl+"kuaixun/", config.Cfg.BiWorldUrl)
	if err != nil {
		logger.Errorf("init bWorld error: %v", err)
		return err
	}
	p.bWorld = bWorld

	tokenViewEn, err := spider.NewSpider(mysql, config.Cfg.TokenviewUrl+"v2api/news/list/all/{{time}}/en", config.Cfg.TokenviewUrl)
	if err != nil {
		logger.Errorf("init tokenViewEn error: %v", err)
		return err
	}
	p.tokenViewEn = tokenViewEn

	tokenViewCn, err := spider.NewSpider(mysql, config.Cfg.TokenviewUrl+"v2api/news/list/all/{{time}}/cn", config.Cfg.TokenviewUrl)
	if err != nil {
		logger.Errorf("init tokenViewCn error: %v", err)
		return err
	}
	p.tokenViewCn = tokenViewCn

	bBtc, err := spider.NewSpider(mysql, config.Cfg.BBtcUrl+"flash", config.Cfg.BBtcUrl)
	if err != nil {
		logger.Errorf("init bBtc error: %v", err)
		return err
	}
	p.bBtc = bBtc

	logger.Info("program inited")
	return nil
}

func (p *program) Start() error {

	ch := make(chan types.Information)
	defer close(ch)
	go p.bWorld.SpiderBishijie(ch)
	go p.tokenViewCn.SpiderTokenview(ch)
	go p.tokenViewEn.SpiderTokenview(ch)
	go p.bBtc.Spider8btc(ch)
	for {
		select {
		case data := <-ch:
			p.bWorld.AddInfodata(data)
		}
	}

	logger.Info("program start")
	return nil
}

func (p *program) Stop() error {

	logger.Info("program stopped")
	return nil
}

func main() {
	cfg := flag.String("C", "config.json", "configuration file")
	flag.Parse()

	if err := config.Init(*cfg); err != nil {
		fmt.Println("init config error:", err.Error())
		return
	}

	app := &program{}
	if err := svc.Run(app); err != nil {
		logger.Println(err)
	}
}
