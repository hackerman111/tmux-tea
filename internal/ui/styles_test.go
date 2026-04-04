package ui

import (
	"strings"
	"testing"
)

func TestPaletteColorsConfigured(t *testing.T) {
	if string(ColorPrimary) == "" {
		t.Fatal("ColorPrimary should not be empty")
	}
	if string(ColorSecondary) == "" {
		t.Fatal("ColorSecondary should not be empty")
	}
	if string(ColorAccent) == "" {
		t.Fatal("ColorAccent should not be empty")
	}
	if string(ColorMuted) == "" {
		t.Fatal("ColorMuted should not be empty")
	}
	if string(ColorDanger) == "" {
		t.Fatal("ColorDanger should not be empty")
	}
}

func TestStylesRenderContent(t *testing.T) {
	title := TitleStyle.Render("Tea")
	if !strings.Contains(title, "Tea") {
		t.Fatalf("TitleStyle should render content, got %q", title)
	}

	selected := SelectedStyle.Render("> Tea")
	if !strings.Contains(selected, "> Tea") {
		t.Fatalf("SelectedStyle should render content, got %q", selected)
	}

	bordered := BorderStyle.Render("Tea")
	if !strings.Contains(bordered, "Tea") {
		t.Fatalf("BorderStyle should render content, got %q", bordered)
	}

	notify := NotifyStyle.Render("READY")
	if !strings.Contains(notify, "READY") {
		t.Fatalf("NotifyStyle should render content, got %q", notify)
	}
}

func TestMenuPopupDimensionsAllowLargerLayouts(t *testing.T) {
	if menuPopupWidth < 72 {
		t.Fatalf("menuPopupWidth = %d, want at least 72", menuPopupWidth)
	}
	if menuPopupHeight < 28 {
		t.Fatalf("menuPopupHeight = %d, want at least 28", menuPopupHeight)
	}
	if panelBodyWidth < 60 {
		t.Fatalf("panelBodyWidth = %d, want at least 60", panelBodyWidth)
	}
}
