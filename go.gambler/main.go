package main

import (
	"fmt"
	"log"
	"main/scheduler"
	"main/scheduler/onebot"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Clip []bool

func (c Clip) String() string {
	s := "["
	for _, b := range c {
		if b {
			s += "●"
		} else {
			s += "○"
		}
	}
	s += "]"
	return s
}

type Player struct {
	ID   int64
	Name string
}

type Game struct {

	// 子弹数量
	Bullet int

	// 弹夹以及子弹序列
	Clip Clip

	// 玩家，首位玩家为房主
	Players []Player

	// 开局状态
	Started bool

	// 开枪轮次
	Round int
}

var players = make([]Player, 0, 100)
var playersMX sync.Mutex
var groupGameList = make(map[string]Game)
var BattleGameList = make(map[string]Game)
var battleRoomIDs = make(map[int64]string)

// RefreshGame 新游戏
// bullets 为子弹数量，capacity 为弹夹容量
func (g *Game) RefreshGame(bullets, capacity int) {
	g.Started = true
	g.Bullet = bullets
	g.Clip = make([]bool, capacity, capacity)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < bullets; i++ {
		for {
			index := r.Intn(capacity)
			if !g.Clip[index] {
				g.Clip[index] = true
				break
			}
		}
	}
	log.Print(g.Clip)
}

func (g *Game) GetRoomID() string {
	ids := make([]string, 0, len(g.Players))
	for _, player := range g.Players {
		ids = append(ids, strconv.FormatInt(player.ID, 10))
	}
	return strings.Join(ids, "-")
}
func (g *Game) WhoseTrun() string {
	player := g.Players[g.Round%2]
	livePercent := float64(len(g.Clip)-g.Bullet-g.Round) / float64(len(g.Clip)-g.Round) * 100
	return fmt.Sprintf("现在轮到 %s 开枪，当前存活概率为 %.2f%%", player.Name, livePercent)
}

func main() {
	s := scheduler.New()
	s.Bind("俄罗斯轮盘", func(c *scheduler.Context) {
		player := Player{
			ID:   c.GetSenderId(),
			Name: c.GetSenderNickname(),
		}

		if c.GetGroupEvent() == nil {
			// 私聊
			if _, ok := battleRoomIDs[player.ID]; ok {
				log.Println(players)
				_, _ = c.Reply("您已正在游戏当中，请勿逃避哦!")
				return
			}

			playersMX.Lock()
			// 查询是否存在正在等待的玩家
			if len(players) > 0 {
				if players[0].ID == player.ID {
					_, _ = c.Reply("您已加入匹配队列，请耐心等候")
					playersMX.Unlock()
					return
				}
				// 匹配到玩家
				game := Game{}
				game.RefreshGame(1, 6)
				game.Players = append(game.Players, players[0], player)

				// 两人的记录分别保存
				roomID := game.GetRoomID()
				battleRoomIDs[game.Players[0].ID] = roomID
				battleRoomIDs[game.Players[1].ID] = roomID
				BattleGameList[roomID] = game

				players = players[1:]
				_, _ = c.GetBot().SendPrivateMessage(
					game.Players[0].ID,
					fmt.Sprintf("已为您匹配到对手：%s\n", game.Players[1].Name)+game.WhoseTrun(),
				)
				_, _ = c.GetBot().SendPrivateMessage(
					game.Players[1].ID,
					fmt.Sprintf("已为您匹配到对手：%s\n", game.Players[0].Name)+game.WhoseTrun(),
				)
			} else {
				players = append(players, player)
				_, _ = c.Reply("正在为您匹配对手，请稍后...")
			}
			playersMX.Unlock()
		} else {
			// 群聊
			game := Game{}
			game.RefreshGame(1, 6)
			game.Players = append(game.Players, player)

			groupId := c.GetGroupId()
			groupGameList[strconv.FormatInt(groupId, 10)] = game
		}
	})
	s.Bind("接受挑战", func(c *scheduler.Context) {})
	s.Bind("开枪", func(c *scheduler.Context) {

		playerID := c.GetSenderId()

		if c.GetGroupEvent() == nil {
			// 私聊

			// 玩家是否已参与游戏
			room, ok := battleRoomIDs[playerID]
			if !ok {
				_, _ = c.Reply("您尚未加入游戏，请输入「俄罗斯转盘」参与")
				return
			}

			// 获取对局
			game, ok := BattleGameList[room] // 这里应该做一个对局存在状态判断
			if !ok || !game.Started {
				log.Printf("获取对局异常，roomID=%v", room)
				return
			}

			// 确认玩家身份
			player := game.Players[game.Round%2]
			opponent := game.Players[(game.Round+1)%2]

			// 是否轮到玩家开枪
			if game.Players[game.Round%2].ID != playerID {
				_, _ = c.Reply(fmt.Sprintf("当前为 %s 的轮次，请耐心等待", player.Name))
				return
			}

			if game.Clip[game.Round] {
				_, _ = c.Reply("你开枪了... Bang! 你死啦！\n游戏结束~")
				_, _ = c.GetBot().SendPrivateMessage(
					opponent.ID,
					fmt.Sprintf(
						"%s 开枪了... Bang! %[1]s 死啦！\n游戏结束~",
						player.Name,
					),
				)
				game.Started = false
				delete(battleRoomIDs, player.ID)
				delete(battleRoomIDs, opponent.ID)
				delete(BattleGameList, room)
			} else {
				game.Round = game.Round + 1
				_, _ = c.Reply("哇！你活下来啦！\n" + game.WhoseTrun())
				_, _ = c.GetBot().SendPrivateMessage(
					opponent.ID,
					fmt.Sprintf(
						"哇！%s 活下来啦！\n",
						player.Name,
					)+game.WhoseTrun(),
				)
				BattleGameList[room] = game
			}
			return
		}

		game, ok := groupGameList[strconv.FormatInt(c.GetGroupId(), 10)]
		if !ok || !game.Started {
			_, _ = c.Reply("当前没有正在进行的游戏")
			return
		}

		if game.Clip[game.Round] {
			_, _ = c.Reply("Bang! 你死啦！")
			game.Started = false
		} else {
			_, _ = c.Reply("哇！你活下来啦！")
			game.Round = game.Round + 1
		}
		groupGameList[strconv.FormatInt(c.GetGroupId(), 10)] = game

	})

	_ = s.Serve(":5800", "/", &onebot.Handler{})
}
