package admin

import (
	"encoding/base64"
	"errors"
)

// GetOps 返回总控快照(带锁)
func (st *Store) GetOps() Ops {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.Ops
}

// FeatureByKey 按 key 查功能(带锁)
func (st *Store) FeatureByKey(key string) (Feature, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, f := range st.Features {
		if f.Key == key {
			return f, true
		}
	}
	return Feature{}, false
}

// FeaturesAll 返回功能快照(带锁, 拷贝)
func (st *Store) FeaturesAll() []Feature {
	st.mu.Lock()
	defer st.mu.Unlock()
	out := make([]Feature, len(st.Features))
	copy(out, st.Features)
	return out
}

// ResolveDefaultProfile 解析默认Profile并解密密钥(每次从盘重读, 管理端改动即时生效)
func ResolveDefaultProfile(dir string) (Profile, string, error) {
	st, err := OpenStore(dir)
	if err != nil {
		return Profile{}, "", err
	}
	st.mu.Lock()
	var sel *Profile
	if st.Ops.DefaultProfileID != "" {
		for i := range st.Profiles {
			if st.Profiles[i].ID == st.Ops.DefaultProfileID && st.Profiles[i].Enabled {
				sel = &st.Profiles[i]
				break
			}
		}
	}
	if sel == nil {
		for i := range st.Profiles {
			if st.Profiles[i].Enabled {
				sel = &st.Profiles[i]
				break
			}
		}
	}
	if sel == nil {
		st.mu.Unlock()
		return Profile{}, "", errors.New("无启用的 Profile, 请在后台 LLM 密钥页配置")
	}
	prof := *sel
	enc := sel.APIKeyEnc
	st.mu.Unlock()
	key, err := decryptKey(dir, enc)
	if err != nil {
		return Profile{}, "", err
	}
	return prof, key, nil
}

// decryptKey 用 secret.key 派生密钥解密(AES-256-GCM, 与 encKey 配对)
func decryptKey(dir, enc string) (string, error) {
	if enc == "" {
		return "", errors.New("API Key 为空")
	}
	aead, err := loadAEAD(dir)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", errors.New("密钥密文非法")
	}
	ns := aead.NonceSize()
	if len(raw) < ns {
		return "", errors.New("密钥密文过短")
	}
	pt, err := aead.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return "", errors.New("密钥解密失败")
	}
	return string(pt), nil
}
