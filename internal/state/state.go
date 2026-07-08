package state

import (
	"sync"

	"smart-cabinet/internal/model"
)

// CabinetState 单柜状态管理（线程安全）
type CabinetState struct {
	mu       sync.RWMutex
	status   model.CabinetStatus
	doorOpen bool // 门是否打开（来自MCU上报的最新状态）
	itemID   int64 // 当前物品ID（0表示无物品）
}

var (
	instance *CabinetState
	once     sync.Once
)

// Get 获取单例
func Get() *CabinetState {
	once.Do(func() {
		instance = &CabinetState{
			status:   model.StatusIdle,
			doorOpen: false,
			itemID:   0,
		}
	})
	return instance
}

// Status 获取当前状态
func (s *CabinetState) Status() model.CabinetStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// DoorOpen 门是否打开
func (s *CabinetState) DoorOpen() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doorOpen
}

// HasItem 是否有物品
func (s *CabinetState) HasItem() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.itemID > 0
}

// ItemID 获取当前物品ID
func (s *CabinetState) ItemID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.itemID
}

// Snapshot 获取状态快照
func (s *CabinetState) Snapshot() model.StatusData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return model.StatusData{
		Status:     s.status,
		StatusText: s.status.Text(),
		HasItem:    s.itemID > 0,
		DoorOpen:   s.doorOpen,
	}
}

// CanStore 是否可以存入
func (s *CabinetState) CanStore() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status == model.StatusIdle
}

// CanRetrieve 是否可以取出
func (s *CabinetState) CanRetrieve() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status == model.StatusOccupied
}

// ---- 状态转换 ----

// StartStore 开始存入操作：空闲 → 等待关门
func (s *CabinetState) StartStore(itemID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = model.StatusWaitingClose
	s.itemID = itemID
	// doorOpen 由 MCU 上报来控制，这里不直接改
}

// StartRetrieve 开始取出操作：已存物 → 等待关门
func (s *CabinetState) StartRetrieve() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = model.StatusWaitingClose
}

// OnDoorClosed MCU 上报门已关闭
func (s *CabinetState) OnDoorClosed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.doorOpen = false

	switch s.status {
	case model.StatusWaitingClose:
		// 等待关门状态下，门关了
		if s.itemID > 0 {
			// 之前有物品 → 变成已存物
			s.status = model.StatusOccupied
		} else {
			// 物品已被取走 → 变成空闲
			s.status = model.StatusIdle
		}
	}
}

// OnDoorOpened MCU 上报门已打开
func (s *CabinetState) OnDoorOpened() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.doorOpen = true
}

// OnItemRetrieved 物品被取走，重置
func (s *CabinetState) OnItemRetrieved() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.itemID = 0
	// 状态保持 waiting_close 直到 MCU 上报关门
}

// SetMCUOnline 标记 MCU 在线状态（预留扩展）
func (s *CabinetState) SetMCUOnline(online bool) {
	// 预留：后续可用来标记 MCU 连接状态
	_ = online
}
