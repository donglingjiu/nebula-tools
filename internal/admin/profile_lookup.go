package admin

// OPT-012: flow 引擎按 ID 解析/校验 profile; 用户侧安全列表(仅 id/name/model)

import "errors"

// ResolveProfileByID 按 ID 解析启用的 Profile 并解密 key(锁模式同 ResolveDefaultProfile)
func ResolveProfileByID(dir, id string) (Profile, string, error) {
	st, err := OpenStore(dir)
	if err != nil {
		return Profile{}, "", err
	}
	st.mu.Lock()
	var sel *Profile
	for i := range st.Profiles {
		if st.Profiles[i].ID == id && st.Profiles[i].Enabled {
			sel = &st.Profiles[i]
			break
		}
	}
	if sel == nil {
		st.mu.Unlock()
		return Profile{}, "", errors.New("profile 不存在或未启用")
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

// ProfileAvailable 校验 id 是否为启用中的 profile
func ProfileAvailable(dir, id string) bool {
	st, err := OpenStore(dir)
	if err != nil {
		return false
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	for i := range st.Profiles {
		if st.Profiles[i].ID == id && st.Profiles[i].Enabled {
			return true
		}
	}
	return false
}

// PubProfile 用户侧可见字段(不含 base_url/api_key)
type PubProfile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Model string `json:"model"`
}

// ListEnabledProfiles 启用 profile 的安全列表
func ListEnabledProfiles(dir string) []PubProfile {
	st, err := OpenStore(dir)
	if err != nil {
		return nil
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	out := []PubProfile{}
	for i := range st.Profiles {
		if st.Profiles[i].Enabled {
			out = append(out, PubProfile{
				ID: st.Profiles[i].ID, Name: st.Profiles[i].Name, Model: st.Profiles[i].Model})
		}
	}
	return out
}
