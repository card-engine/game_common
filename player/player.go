package player

import "sync"

// Deprecated: The functionality in this file is deprecated.
type Player struct {
	AId          uint64  // 用户ID
	PlayerId     string  // 玩家ID
	Username     string  // 玩家用户名
	ProfileImage string  // 玩家头像
	Balance      float64 // 玩家余额
	AppId        string  // 玩家AppId
	mu           sync.Mutex

	Extra sync.Map // 并发安全的扩展字段
}

// Deprecated: The functionality in this file is deprecated.
func NewPlayer(playerId string, balance float64) *Player {
	return &Player{}
}

// Deprecated: The functionality in this file is deprecated.
// 设置自定义指针
func (p *Player) SetExtra(key string, value interface{}) {
	p.Extra.Store(key, value)
}

// Deprecated: The functionality in this file is deprecated.
// 获取自定义指针
func (p *Player) GetExtra(key string) interface{} {
	val, _ := p.Extra.Load(key)
	return val
}

// Deprecated: The functionality in this file is deprecated.
func (p *Player) RemoveExtra(key string) {
	p.Extra.Delete(key)
}

// Deprecated: The functionality in this file is deprecated.
// 扣钱，线程安全
func (p *Player) MinusMoney(amount float64) bool {
	if amount <= 0 {
		return false
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Balance < amount {
		return false
	}
	p.Balance -= amount
	return true
}

// Deprecated: The functionality in this file is deprecated.
func (p *Player) AddMoney(amount float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Balance += amount
}

// Deprecated: The functionality in this file is deprecated.
func (p *Player) GetBalance() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Balance
}
