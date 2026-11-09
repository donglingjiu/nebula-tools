package ai

// OPT-016a: 材料库存储层（自包含; 路由注册与注入接线在 OPT-016b）
// 设计: 一材料一 JSON 落 s.dir/materials/, 全量保存不截断——截断只应发生在注入预算层
import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Material 一份可跨流程复用的参考资料
type Material struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const matMaxBytes = 2 << 20 // 单条 2MB 上限

func (s *Service) matDir() string { return filepath.Join(s.dir, "materials") }

func matSafeID(id string) string {
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		}
	}
	return b.String()
}

func newMatID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(b)
}

// SaveMaterial 新建材料（全量落盘）
func (s *Service) SaveMaterial(title, content string) (*Material, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" || content == "" {
		return nil, errors.New("title 与 content 必填")
	}
	if len(content) > matMaxBytes {
		return nil, errors.New("材料超过 2MB 上限")
	}
	if err := os.MkdirAll(s.matDir(), 0o755); err != nil {
		return nil, err
	}
	now := time.Now()
	m := &Material{ID: newMatID(), Title: title, Content: content, CreatedAt: now, UpdatedAt: now}
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(s.matDir(), m.ID+".json"), raw, 0o644); err != nil {
		return nil, err
	}
	return m, nil
}

// ListMaterials 全量列表（按更新时间倒序）
func (s *Service) ListMaterials() ([]Material, error) {
	ents, err := os.ReadDir(s.matDir())
	if err != nil {
		if os.IsNotExist(err) {
			return []Material{}, nil
		}
		return nil, err
	}
	out := make([]Material, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(s.matDir(), e.Name()))
		if err != nil {
			continue
		}
		var m Material
		if json.Unmarshal(raw, &m) == nil && m.ID != "" {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

// GetMaterial 按 ID 取回（含全文）
func (s *Service) GetMaterial(id string) (*Material, error) {
	id = matSafeID(id)
	if id == "" {
		return nil, errors.New("非法材料 id")
	}
	raw, err := os.ReadFile(filepath.Join(s.matDir(), id+".json"))
	if err != nil {
		return nil, errors.New("材料不存在")
	}
	var m Material
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, errors.New("材料文件损坏")
	}
	return &m, nil
}

// DeleteMaterial 删除
func (s *Service) DeleteMaterial(id string) error {
	id = matSafeID(id)
	if id == "" {
		return errors.New("非法材料 id")
	}
	if err := os.Remove(filepath.Join(s.matDir(), id+".json")); err != nil {
		if os.IsNotExist(err) {
			return errors.New("材料不存在")
		}
		return err
	}
	return nil
}
