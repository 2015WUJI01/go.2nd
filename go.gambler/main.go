package main

import (
	"fmt"
	"log"
	"main/scheduler"
	"main/scheduler/onebot"
	"sync"
	"time"
)

// 匹配队列
var matchQueue = make([]Player, 0, 100)
var queueMx sync.Mutex

// 群聊游戏房间
var gameRooms = make(map[string]*Game)

// 玩家与房间对应关系
var playerToRoom = make(map[int64]string)
var roomMx sync.Mutex

var cleanSignal = make(chan struct{})

// MatchGame 匹配游戏
func MatchGame() {
	if len(matchQueue) >= 2 {
		queueMx.Lock()
		game := NewGame(1, 7)
		game.Type = gamePrivate
		game.Players = matchQueue[:2]
		matchQueue = matchQueue[2:]
		queueMx.Unlock()

		roomMx.Lock()
		roomId := fmt.Sprintf("p%d", game.Players[0].ID)
		gameRooms[roomId] = &game
		for _, p := range game.Players {
			playerToRoom[p.ID] = roomId
			msg := fmt.Sprintf("已匹配到对手，游戏开始。\n%s", game.WhoseTurn())
			_, _ = p.Ctx.Reply(msg)
		}
		roomMx.Unlock()
	}
}

// CleanGame 清理超时的游戏与已结束的游戏
func CleanGame() {
	roomMx.Lock()
	defer roomMx.Unlock()
	for grk, game := range gameRooms {
		if game.State == gameEnd {
			delete(gameRooms, grk)
			for _, p := range game.Players {
				delete(playerToRoom, p.ID)
			}
		} else {
			if time.Now().Sub(game.UpdatedAt) > 75*time.Second {
				delete(gameRooms, grk)
				if game.Type == gamePrivate {
					for _, p := range game.Players {
						delete(playerToRoom, p.ID)
						_, _ = p.Ctx.Reply("玩家太长时间未操作，游戏已自动结束")
					}
				} else {

				}
			}
		}
	}
}

// Roulette 轮盘赌准备开局
func Roulette(c *scheduler.Context) {
	if c.IsPrivate() {
		roomMx.Lock()
		if room, ok := playerToRoom[c.GetSenderId()]; ok && gameRooms[room].State != gameEnd {
			_, _ = c.Reply("您已在游戏房间中, 请勿重复开局！")
			roomMx.Unlock()
			return
		}
		roomMx.Unlock()

		queueMx.Lock()
		for _, player := range matchQueue {
			if player.ID == c.GetSenderId() {
				_, _ = c.Reply("您当前正在匹配队列中！")
				queueMx.Unlock()
				return
			}
		}
		queueMx.Unlock()

		_, _ = c.Reply("正在为您匹配对手，请耐心等候...")
		queueMx.Lock()
		matchQueue = append(matchQueue, Player{
			ID:   c.GetSenderId(),
			Name: c.GetSenderNickname(),
			Ctx:  c,
		})
		queueMx.Unlock()

	} else if c.IsGroup() {

	} else {

	}
}

// Shoot 开枪
func Shoot(c *scheduler.Context) {
	if c.IsPrivate() {
		roomMx.Lock()
		room, ok := playerToRoom[c.GetSenderId()]
		if !ok || ok && gameRooms[room].State == gameEnd {
			_, _ = c.Reply("您当前还不在对局中哦~")
			roomMx.Unlock()
			return
		}
		game := gameRooms[room]
		roomMx.Unlock()

		if game.GetRoundPlayer().ID != c.GetSenderId() {
			_, _ = c.Reply("还没轮到你呢，这么急着送死吗？")
			return
		}

		if game.Clip[game.Round] {
			msg := "黑洞洞的枪口冒出了火焰，一颗子弹送走了%s"
			game.Round++
			_, _ = c.Reply(fmt.Sprintf(msg, "你"))
			_, _ = game.GetRoundPlayer().Ctx.Reply(fmt.Sprintf(msg, c.GetSenderNickname()))
			game.State = gameEnd
			cleanSignal <- struct{}{}
			return
		} else {
			msgNotHappen := "%s轻轻扣动扳机，咔哒一声，什么也没有发生"
			game.UpdatedAt = time.Now()
			game.Round++
			_, _ = c.Reply(fmt.Sprintf(msgNotHappen, "你") + "\n" + game.WhoseTurn())
			_, _ = game.GetRoundPlayer().Ctx.Reply(fmt.Sprintf(msgNotHappen, c.GetSenderNickname()) + "\n" + game.WhoseTurn())
			return
		}

	}

}

func main() {
	s := scheduler.New()
	s.Bind("俄罗斯轮盘", Roulette)
	s.Bind("接受挑战", func(c *scheduler.Context) {})
	s.Bind("开枪", Shoot)

	// 每200ms做一次匹配
	go func() {
		t := time.Tick(200 * time.Millisecond)
		for {
			<-t
			MatchGame()
		}
	}()

	// 每500ms清理一次游戏
	go func() {
		t := time.Tick(20000 * time.Millisecond)
		for {
			select {
			case <-t:
			case <-cleanSignal:
			}
			log.Println("清理一次游戏")
			CleanGame()
		}
	}()

	go func() {
		t := time.Tick(2000 * time.Millisecond)
		for {
			<-t
			queueMx.Lock()
			roomMx.Lock()
			waitPlayers := make([]string, 0)
			for _, p := range matchQueue {
				waitPlayers = append(waitPlayers, p.Name)
			}
			log.Println("待匹配玩家:", waitPlayers)

			rooms := make([]string, 0)
			for room := range gameRooms {
				rooms = append(rooms, room)
			}
			log.Println("当前游戏房间:", rooms)

			players := make([]string, 0)
			for player, room := range playerToRoom {
				players = append(players, fmt.Sprintf("%d-%s", player, room))
			}
			log.Println("当前游戏中的玩家:", players)

			log.Println("--------------------------------------------")

			roomMx.Unlock()
			queueMx.Unlock()
		}

	}()

	_ = s.Serve(":5800", "/", &onebot.Handler{})
}
