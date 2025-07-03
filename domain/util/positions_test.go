package util

import (
	"testing"

	"github.com/VladMinzatu/reference-manager/domain/model"
)

func TestValidatePositions(t *testing.T) {
	ids := []model.Id{1, 2, 3}

	t.Run("valid permutation", func(t *testing.T) {
		positions := map[model.Id]int{1: 2, 2: 0, 3: 1}
		err := ValidatePositions(ids, positions)
		if err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}
	})

	t.Run("missing id", func(t *testing.T) {
		positions := map[model.Id]int{1: 0, 2: 1}
		err := ValidatePositions(ids, positions)
		if err == nil || err.Error() != "positions map must have exactly 3 entries" {
			t.Errorf("expected error for missing id, got: %v", err)
		}
	})

	t.Run("duplicate position", func(t *testing.T) {
		positions := map[model.Id]int{1: 0, 2: 1, 3: 1}
		err := ValidatePositions(ids, positions)
		if err == nil || err.Error() != "duplicate position 1" {
			t.Errorf("expected error for duplicate position, got: %v", err)
		}
	})

	t.Run("out of range position", func(t *testing.T) {
		positions := map[model.Id]int{1: 0, 2: 1, 3: 3}
		err := ValidatePositions(ids, positions)
		if err == nil || err.Error() != "invalid position 3 for id 3" {
			t.Errorf("expected error for out of range, got: %v", err)
		}
	})

	t.Run("missing position for id", func(t *testing.T) {
		positions := map[model.Id]int{1: 0, 2: 1, 4: 2}
		err := ValidatePositions(ids, positions)
		if err == nil || err.Error() != "missing position for id 3" {
			t.Errorf("expected error for missing position for id, got: %v", err)
		}
	})
}
