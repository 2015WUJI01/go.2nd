package main

import (
	"fmt"
	"main/scheduler"
	"main/scheduler/onebot"
	"time"
)

type Follower struct {
	ID   int64
	Name string
	Ctx  *scheduler.Context
}

var followers []Follower

func UpdateNewsAndSendMsg() {
	link, content := Fetch()
	if readLink() != link {
		go writeLink(link)
		msg := fmt.Sprintf("【公告】%s\n 原文地址：%s", content, link)
		for _, follower := range followers {
			_, _ = follower.Ctx.Reply(msg)
		}
	}
}

func main() {

	s := scheduler.New()
	s.Bind("公告", News).Alias("最新公告")
	s.Bind("订阅", Subscribe).Alias("订阅公告")
	s.Bind("取消订阅", Unsubscribe).Alias("取消订阅公告")

	go func() {
		t := time.Tick(2 * time.Minute)
		for {
			<-t
			UpdateNewsAndSendMsg()
		}
	}()

	_ = s.Serve(":9802", "/", &onebot.Handler{})
}
