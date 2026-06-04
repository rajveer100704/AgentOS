package audit

import "fmt"

type VerifyResult struct {
	Valid        bool   `json:"valid"`
	TotalEntries int    `json:"total_entries"`
	ErrorAt      int64  `json:"error_at,omitempty"`
	Message      string `json:"message"`
}

func (l *Logger) Verify() (VerifyResult, error) {
	entries, err := l.store.Query(QueryFilters{Limit: 0}) // 0 = all
	if err != nil {
		return VerifyResult{}, err
	}
	if len(entries) == 0 {
		return VerifyResult{Valid: true, Message: "audit log is empty"}, nil
	}

	for i, entry := range entries {
		expected := computeHash(entry)
		if expected != entry.EntryHash {
			return VerifyResult{
				Valid:        false,
				TotalEntries: len(entries),
				ErrorAt:      entry.ID,
				Message:      fmt.Sprintf("hash mismatch at entry %d", entry.ID),
			}, nil
		}
		if i > 0 && entry.PreviousHash != entries[i-1].EntryHash {
			return VerifyResult{
				Valid:        false,
				TotalEntries: len(entries),
				ErrorAt:      entry.ID,
				Message:      fmt.Sprintf("chain break at entry %d", entry.ID),
			}, nil
		}
	}

	return VerifyResult{
		Valid:        true,
		TotalEntries: len(entries),
		Message:      "audit log integrity verified",
	}, nil
}
