package main

import (
	"fmt"
	"main/scheduler"
)

func News(c *scheduler.Context) {
	link, content := Fetch()
	msg := fmt.Sprintf("【公告】%s\n 原文地址：%s", content, link)
	if readLink() != link {
		go writeLink(link)
	}
	_, _ = c.Reply(msg)
}

func Subscribe(c *scheduler.Context) {
	id := c.GetSenderId()
	Subscribed := false
	for _, follower := range followers {
		if follower.ID == id {
			Subscribed = true
			_, _ = c.Reply("您已订阅，无需重复订阅。若需取消订阅请私聊发送「取消订阅」。")
		}
	}

	if !Subscribed {
		followers = append(followers, Follower{
			ID:   id,
			Name: c.GetSenderNickname(),
			Ctx:  c,
		})
		_, _ = c.Reply("订阅成功！我们将在最新的公告发布时私聊通知您，如需手动获取最新一条公告，请尝试私聊发送「公告」或「最新公告」。")
	}
}

func Unsubscribe(c *scheduler.Context) {
	id := c.GetSenderId()
	unsubscribed := true
	for idx, follower := range followers {
		if follower.ID == id {
			unsubscribed = false
			followers = append(followers[:idx], followers[idx+1:]...)
			_, _ = c.Reply("订阅成功！我们将在最新的公告发布时私聊通知您，如需手动获取最新一条公告，请尝试私聊发送「公告」或「最新公告」。")
		}
	}

	if unsubscribed {
		_, _ = c.Reply("您已取消订阅。若需订阅请私聊发送「订阅」或「订阅公告」。")
	}
}
