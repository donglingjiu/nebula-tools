package ai

// OPT-029: GET /api/ai/skills — embedded base skill-pack contracts.
import (
	"encoding/json"
	"net/http"

	"nova/internal/skills"
)

func (s *Service) skillsList(w http.ResponseWriter, r *http.Request) {
	packs, err := skills.Load()
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "skills load failed: " + err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]any{"packs": packs, "count": len(packs)})
}
