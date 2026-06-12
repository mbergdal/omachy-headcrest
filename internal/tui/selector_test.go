package tui

import "testing"

func TestPackageSelectorSelectedNames(t *testing.T) {
	selector := NewPackageSelectorModel([]PackageChoice{
		{Name: "ghostty", Selected: true},
		{Name: "zed"},
		{Name: "mise", Selected: true},
	})

	selected := selector.SelectedNames()
	if len(selected) != 2 || selected[0] != "ghostty" || selected[1] != "mise" {
		t.Fatalf("SelectedNames() = %v, want [ghostty mise]", selected)
	}
}

func TestPackageSelectorSelectedNamesReturnsEmptySlice(t *testing.T) {
	selector := NewPackageSelectorModel([]PackageChoice{{Name: "ghostty"}})
	selected := selector.SelectedNames()
	if selected == nil {
		t.Fatal("SelectedNames should return an empty slice, not nil")
	}
	if len(selected) != 0 {
		t.Fatalf("SelectedNames() = %v, want empty", selected)
	}
}

func TestPackageSelectorToggleAndBulkActions(t *testing.T) {
	selector := NewPackageSelectorModel([]PackageChoice{
		{Name: "ghostty"},
		{Name: "zed", Selected: true},
	})

	selector.Toggle()
	if !selector.Items[0].Selected {
		t.Fatal("Toggle should select current item")
	}

	selector.SelectAll(false)
	if selector.SelectedCount() != 0 {
		t.Fatalf("SelectAll(false) selected %d items", selector.SelectedCount())
	}

	selector.Invert()
	if selector.SelectedCount() != 2 {
		t.Fatalf("Invert selected %d items, want 2", selector.SelectedCount())
	}
}

func TestPackageSelectorMoveBounds(t *testing.T) {
	selector := NewPackageSelectorModel([]PackageChoice{
		{Name: "one"},
		{Name: "two"},
	})

	selector.Move(-10, 20)
	if selector.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", selector.cursor)
	}

	selector.Move(10, 20)
	if selector.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", selector.cursor)
	}
}
