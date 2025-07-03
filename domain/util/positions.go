package util

import (
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
)

// Checks that the positions map is a permutation of 0..n-1 for the given ids.
func ValidatePositions(ids []model.Id, positions map[model.Id]int) error {
	if len(ids) != len(positions) {
		return fmt.Errorf("positions map must have exactly %d entries", len(ids))
	}
	seen := make(map[int]bool, len(ids))
	for _, id := range ids {
		pos, ok := positions[id]
		if !ok {
			return fmt.Errorf("missing position for id %v", id)
		}
		if pos < 0 || pos >= len(ids) {
			return fmt.Errorf("invalid position %d for id %v", pos, id)
		}
		if seen[pos] {
			return fmt.Errorf("duplicate position %d", pos)
		}
		seen[pos] = true
	}
	return nil
}
