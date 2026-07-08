package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"smart-cabinet/internal/model"
	"smart-cabinet/internal/service"
)

// CabinetHandler HTTP 接口处理器
type CabinetHandler struct {
	svc *service.CabinetService
}

// NewCabinetHandler 创建处理器
func NewCabinetHandler(svc *service.CabinetService) *CabinetHandler {
	return &CabinetHandler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *CabinetHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/cabinet/store", h.handleStore)
	mux.HandleFunc("/api/v1/cabinet/retrieve", h.handleRetrieve)
	mux.HandleFunc("/api/v1/cabinet/status", h.handleStatus)
}

// POST /api/v1/cabinet/store
func (h *CabinetHandler) handleStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.Response{
			Code: 405, Message: "仅支持 POST 方法",
		})
		return
	}

	var req model.StoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.Response{
			Code: 40000, Message: "请求格式错误",
		})
		return
	}

	err := h.svc.Store(req.Code)
	if err != nil {
		code, msg := mapError(err)
		writeJSON(w, http.StatusOK, model.Response{Code: code, Message: msg})
		return
	}

	log.Printf("[API] 存入请求成功，密码: %s", req.Code)
	writeJSON(w, http.StatusOK, model.Response{
		Code: 0, Message: "柜门已打开，请放入外卖后关门",
	})
}

// POST /api/v1/cabinet/retrieve
func (h *CabinetHandler) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.Response{
			Code: 405, Message: "仅支持 POST 方法",
		})
		return
	}

	var req model.RetrieveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.Response{
			Code: 40000, Message: "请求格式错误",
		})
		return
	}

	err := h.svc.Retrieve(req.Code)
	if err != nil {
		code, msg := mapError(err)
		writeJSON(w, http.StatusOK, model.Response{Code: code, Message: msg})
		return
	}

	log.Printf("[API] 取出请求成功，密码: %s", req.Code)
	writeJSON(w, http.StatusOK, model.Response{
		Code: 0, Message: "验证通过，柜门已打开，请取走外卖后关门",
	})
}

// GET /api/v1/cabinet/status
func (h *CabinetHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.Response{
			Code: 405, Message: "仅支持 GET 方法",
		})
		return
	}

	data := h.svc.GetStatus()
	writeJSON(w, http.StatusOK, model.Response{
		Code: 0,
		Data: data,
	})
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// mapError 将业务错误映射为错误码和消息
func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrInvalidCode):
		return 40001, err.Error()
	case errors.Is(err, service.ErrCabinetBusy):
		return 40002, err.Error()
	case errors.Is(err, service.ErrCodeMismatch):
		return 40003, err.Error()
	case errors.Is(err, service.ErrCabinetEmpty):
		return 40004, err.Error()
	case errors.Is(err, service.ErrMCUNotConnected):
		return 50003, err.Error()
	case errors.Is(err, service.ErrMCUSendFailed):
		return 50001, err.Error()
	default:
		return 50002, "服务器内部错误: " + err.Error()
	}
}
