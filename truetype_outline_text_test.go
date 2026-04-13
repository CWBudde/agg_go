//go:build !freetype

package agg

import "testing"

func TestNewFreeTypeOutlineTextRequiresFreetypeTag(t *testing.T) {
	txt, err := NewFreeTypeOutlineText()
	if err == nil {
		if txt != nil {
			_ = txt.Close()
		}
		t.Fatal("expected freetype-tag constructor error")
	}
}
