package model

import "time"

// CabinetStatus 柜子状态枚举
type CabinetStatus string

const (
	StatusIdle           CabinetStatus = "idle"            // 空闲（无物品，门关）
	StatusWaitingClose   CabinetStatus = "waiting_close"   // 等待关门
	StatusOccupied       CabinetStatus = "occupied"        // 已存物，门关
)

// StatusText 返回状态的中文描述
func (s CabinetStatus) Text() string {
	switch s {
	case StatusIdle:
		return "空闲"
	case StatusWaitingClose:
		return "等待关门"
	case StatusOccupied:
		return "已存物"
	default:
		return "未知"
	}
}

// StoreRequest 存入请求
type StoreRequest struct {
	Code string `json:"code"` // 4位数字密码
}

// RetrieveRequest 取出请求
type RetrieveRequest struct {
	Code string `json:"code"` // 4位数字密码
}

// Response 通用响应
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// StatusData 柜子状态数据
type StatusData struct {
	Status     CabinetStatus `json:"status"`
	StatusText string        `json:"status_text"`
	HasItem    bool          `json:"has_item"`
	DoorOpen   bool          `json:"door_open"`
}

// CurrentItem 当前柜内物品
type CurrentItem struct {
	ID       int64     `json:"id"`
	Code     string    `json:"code"`
	StoredAt time.Time `json:"stored_at"`
	Status   string    `json:"status"`
}

// Record 存取记录
type Record struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Action    string    `json:"action"` // store / retrieve
	CreatedAt time.Time `json:"created_at"`
}
