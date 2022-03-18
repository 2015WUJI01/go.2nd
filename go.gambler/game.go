package main

import (
	"fmt"
	"main/scheduler"
	"math/rand"
	"time"
)

type gameState int
type gameType int

const (
	gameReady gameState = iota
	gameStart
	gameEnd
)

const (
	gamePrivate gameType = iota
	gameGroup
)

type Player struct {
	ID   int64
	Name string
	Ctx  *scheduler.Context
}

type Game struct {
	Type gameType

	// 子弹数量
	Bullet int

	// 弹夹以及子弹序列
	Clip []bool

	// 玩家，首位玩家为房主
	Players []Player

	// 开局状态
	State gameState

	// 开枪轮次
	Round int

	// 最后处理时间
	UpdatedAt time.Time
}

func (g Game) FormatClip() string {
	s := "["
	for _, b := range g.Clip {
		if b {
			s += "●"
		} else {
			s += "○"
		}
	}
	s += "]"
	return s
}

func (g Game) CalShotProbability() float64 {
	return float64(g.Bullet) / float64(len(g.Clip)-g.Round) * 100
}

func (g Game) GetRoundPlayer() Player {
	return g.Players[g.Round%len(g.Players)]
}

func (g Game) WhoseTurn() string {
	return fmt.Sprintf("现在轮到 %s 开枪，下一枪的中弹概率为 %.2f%%", g.GetRoundPlayer().Name, g.CalShotProbability())
}

func NewGame(bullets, capacity int) Game {
	if bullets > capacity {
		bullets = capacity
	}
	g := Game{
		Bullet:    bullets,
		Clip:      make([]bool, capacity),
		Players:   make([]Player, 0),
		State:     gameReady,
		Round:     0,
		UpdatedAt: time.Now(),
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < bullets; i++ {
		g.Clip[i] = true
	}
	r.Shuffle(len(g.Clip), func(i, j int) {
		g.Clip[i], g.Clip[j] = g.Clip[j], g.Clip[i]
	})
	return g
}
