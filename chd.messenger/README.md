# chd.messager
彩虹岛公告通知。定时抓取公告页面，如果更新了公告就发送给已订阅的用户。

### 版本信息

- cq-http: [1.0.0-rc1](https://github.com/Mrs4s/go-cqhttp/releases/tag/v1.0.0-rc1)
- go: 1.17
- gocolly/colly: v1.2.0

### 实现过程
使用 [Colly](http://go-colly.org/) 抓取指定页面的信息，然后利用 cqhttp 机器人发送消息给指定用户。
用户可以手动发送指令进行实时获取，也可以发送「订阅」等待机器人自动发送，定时 2 分钟获取一次信息，并将最新的公告链接地址作为标识保存在 latest_news_link 文件中，若公告更新则更新文件，否则只使用文件中的内容。

### 待完善
订阅用户列表用的是 map，重启服务时会丢失，应序列化后存本地文件副本。