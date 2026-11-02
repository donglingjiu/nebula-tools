package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"nova/internal/admin"
	"nova/internal/auth"
)

// Order 模拟订单: status paid 即权益(无第二张权益表)
type Order struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	FeatureKey string  `json:"feature_key"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
	PaidAt     string  `json:"paid_at,omitempty"`
}

type orderStore struct {
	Orders map[string]*Order `json:"orders"`
}

func (s *Service) loadOrders() (*orderStore, error) {
	st := &orderStore{Orders: map[string]*Order{}}
	b, err := os.ReadFile(filepath.Join(s.dir, "orders.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(b, st); err != nil {
		return nil, err
	}
	if st.Orders == nil {
		st.Orders = map[string]*Order{}
	}
	return st, nil
}

func (s *Service) saveOrders(st *orderStore) error {
	b, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		return err
	}
	p := filepath.Join(s.dir, "orders.json")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// hasPaidOrder 权益判定: 该user对该feature有paid订单
func (s *Service) hasPaidOrder(userID, key string) bool {
	st, err := s.loadOrders()
	if err != nil {
		return false
	}
	for _, o := range st.Orders {
		if o.UserID == userID && o.FeatureKey == key && o.Status == "paid" {
			return true
		}
	}
	return false
}

// gateCheck 门链+权益: 402时已购放行, 未购则追加购买指引
func (s *Service) gateCheck(u *auth.User, want string) (*admin.Store, string, *gateErr) {
	st, key, ge := s.resolveFeature(want)
	if ge == nil || ge.code != 402 {
		return st, key, ge
	}
	if u != nil && s.hasPaidOrder(u.ID, key) {
		return st, key, nil
	}
	ge.msg += " | 可购买: POST /api/ai/order"
	return st, key, ge
}

func (s *Service) orderCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FeatureKey string `json:"feature_key"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	u := s.auth.UserFrom(r)
	st, key, ge := s.resolveFeature(req.FeatureKey)
	if ge != nil && ge.code != 402 {
		writeGate(w, ge)
		return
	}
	if ge == nil {
		fail(w, 400, "该功能当前无需购买")
		return
	}
	if st == nil {
		var oerr error
		st, oerr = admin.OpenStore(s.dir)
		if oerr != nil {
			fail(w, 500, "读取配置失败")
			return
		}
	}
	f, ok := st.FeatureByKey(key)
	if !ok || !f.IsPaid {
		fail(w, 400, "功能不可购买")
		return
	}
	ost, err := s.loadOrders()
	if err != nil {
		fail(w, 500, "读取订单失败")
		return
	}
	o := &Order{
		ID:         fmt.Sprintf("%x", time.Now().UnixNano()),
		UserID:     u.ID,
		FeatureKey: key,
		Price:      f.Price,
		Status:     "created",
		CreatedAt:  time.Now().Format(time.RFC3339),
	}
	ost.Orders[o.ID] = o
	if err := s.saveOrders(ost); err != nil {
		fail(w, 500, "保存订单失败")
		return
	}
	writeJSON(w, 200, o)
}

func (s *Service) orderPay(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ID == "" {
		fail(w, 400, "参数错误: 需要 id")
		return
	}
	u := s.auth.UserFrom(r)
	ost, err := s.loadOrders()
	if err != nil {
		fail(w, 500, "读取订单失败")
		return
	}
	o, ok := ost.Orders[req.ID]
	if !ok || o.UserID != u.ID {
		fail(w, 404, "订单不存在")
		return
	}
	if o.Status != "paid" {
		o.Status = "paid"
		o.PaidAt = time.Now().Format(time.RFC3339)
		if err := s.saveOrders(ost); err != nil {
			fail(w, 500, "保存订单失败")
			return
		}
	}
	writeJSON(w, 200, o)
}

func (s *Service) orderList(w http.ResponseWriter, r *http.Request) {
	u := s.auth.UserFrom(r)
	ost, err := s.loadOrders()
	if err != nil {
		fail(w, 500, "读取订单失败")
		return
	}
	mine := []*Order{}
	for _, o := range ost.Orders {
		if o.UserID == u.ID {
			mine = append(mine, o)
		}
	}
	writeJSON(w, 200, map[string]any{"orders": mine})
}
