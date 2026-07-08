package service

import (
	"errors"
	"regexp"

	"smart-cabinet/internal/db"
	"smart-cabinet/internal/mcu"
	"smart-cabinet/internal/model"
	"smart-cabinet/internal/state"
)

var codeRegex = regexp.MustCompile(`^\d{4}$`)

// 业务错误
var (
	ErrInvalidCode     = errors.New("密码必须为4位数字")
	ErrCabinetBusy     = errors.New("柜子当前不可用，请稍后再试")
	ErrCodeMismatch    = errors.New("密码错误，请重试")
	ErrCabinetEmpty    = errors.New("柜子内无物品可取")
	ErrMCUNotConnected = errors.New("单片机不在线")
	ErrMCUSendFailed   = errors.New("与单片机通信失败")
)

// CabinetService 柜子业务服务
type CabinetService struct {
	dbStore   *db.DB
	mcuServer *mcu.Server
	state     *state.CabinetState
}

// NewCabinetService 创建业务服务
func NewCabinetService(database *db.DB, mcuSrv *mcu.Server) *CabinetService {
	svc := &CabinetService{
		dbStore:   database,
		mcuServer: mcuSrv,
		state:     state.Get(),
	}

	// 注册 MCU 回调
	mcuSrv.OnDoorClosed = svc.onDoorClosed
	mcuSrv.OnDoorOpened = svc.onDoorOpened

	return svc
}

// ---- 回调处理 ----

func (s *CabinetService) onDoorClosed() {
	s.state.OnDoorClosed()
}

func (s *CabinetService) onDoorOpened() {
	s.state.OnDoorOpened()
}

// ---- 业务方法 ----

// Store 存入物品
func (s *CabinetService) Store(code string) error {
	// 1. 校验密码格式
	if !codeRegex.MatchString(code) {
		return ErrInvalidCode
	}

	// 2. 检查柜子是否可用
	if !s.state.CanStore() {
		return ErrCabinetBusy
	}

	// 3. 检查 MCU 连接
	if !s.mcuServer.IsConnected() {
		return ErrMCUNotConnected
	}

	// 4. 写入数据库
	if err := s.dbStore.SetCurrentItem(code); err != nil {
		return err
	}
	if err := s.dbStore.InsertRecord(code, "store"); err != nil {
		return err
	}

	// 5. 获取刚插入的物品ID
	item, err := s.dbStore.GetCurrentItem()
	if err != nil || item == nil {
		return errors.New("获取物品信息失败")
	}

	// 6. 更新状态机
	s.state.StartStore(item.ID)

	// 7. 发送开门指令
	if err := s.mcuServer.SendOpen(); err != nil {
		return ErrMCUSendFailed
	}

	return nil
}

// Retrieve 取出物品
func (s *CabinetService) Retrieve(code string) error {
	// 1. 校验密码格式
	if !codeRegex.MatchString(code) {
		return ErrInvalidCode
	}

	// 2. 检查柜子是否有物品
	if !s.state.CanRetrieve() {
		return ErrCabinetEmpty
	}

	// 3. 检查 MCU 连接
	if !s.mcuServer.IsConnected() {
		return ErrMCUNotConnected
	}

	// 4. 校验密码
	ok, err := s.dbStore.VerifyCode(code)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCodeMismatch
	}

	// 5. 获取当前物品
	item, err := s.dbStore.GetCurrentItem()
	if err != nil || item == nil {
		return ErrCabinetEmpty
	}

	// 6. 标记已取走
	if err := s.dbStore.MarkRetrieved(item.ID); err != nil {
		return err
	}
	if err := s.dbStore.InsertRecord(code, "retrieve"); err != nil {
		return err
	}

	// 7. 更新状态机
	s.state.StartRetrieve()
	s.state.OnItemRetrieved()

	// 8. 发送开门指令
	if err := s.mcuServer.SendOpen(); err != nil {
		return ErrMCUSendFailed
	}

	return nil
}

// GetStatus 获取柜子状态
func (s *CabinetService) GetStatus() model.StatusData {
	return s.state.Snapshot()
}
