package tagcolor

import "testing"

func TestColorFromName_Deterministic(t *testing.T) {
	c1 := ColorFromName("Go")
	c2 := ColorFromName("Go")
	if c1 != c2 {
		t.Errorf("same name should produce same color: %s != %s", c1, c2)
	}
}

func TestColorFromName_DifferentNames(t *testing.T) {
	c1 := ColorFromName("Go")
	c2 := ColorFromName("React")
	if c1 == c2 {
		t.Errorf("different names should likely produce different colors: %s == %s", c1, c2)
	}
}

func TestColorFromName_EmptyName(t *testing.T) {
	c := ColorFromName("")
	if c == "" {
		t.Error("empty name should still return a color")
	}
	if c != Palette[0] {
		t.Errorf("empty name should return first palette color, got %s", c)
	}
}

func TestColorFromName_ValidHex(t *testing.T) {
	c := ColorFromName("Test")
	if !IsValidHex(c) {
		t.Errorf("color should be valid 7-char HEX: got %s", c)
	}
}

func TestColorFromName_AllPaletteColors(t *testing.T) {
	seen := make(map[string]bool)
	testNames := []string{
		"Go", "React", "Docker", "Kubernetes", "TypeScript",
		"Python", "AWS", "Vue", "Redis", "GraphQL",
		"MongoDB", "Linux", "Node.js", "Rust", "C++",
		"Java", "Swift", "Flutter", "DevOps", "ML",
	}
	for _, name := range testNames {
		c := ColorFromName(name)
		if !IsValidHex(c) {
			t.Errorf("invalid hex color for %s: %s", name, c)
		}
		seen[c] = true
	}
	if len(seen) < 8 {
		t.Errorf("expected more color variety, got %d unique colors from 20 names", len(seen))
	}
}

func TestColorFromName_ReturnsPaletteColor(t *testing.T) {
	paletteSet := make(map[string]bool)
	for _, c := range Palette {
		paletteSet[c] = true
	}
	testNames := []string{"Go", "React", "Docker", "Test", "Hello", "World"}
	for _, name := range testNames {
		c := ColorFromName(name)
		if !paletteSet[c] {
			t.Errorf("ColorFromName(%q) = %q, not in palette", name, c)
		}
	}
}

func TestIsValidHex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"#ef4444", true},
		{"#000000", true},
		{"#FFFFFF", true},
		{"#123abc", true},
		{"", false},
		{"red", false},
		{"#xyz", false},
		{"#12345", false},
		{"#1234567", false},
		{"123456", false},
		{"#12 345", false},
	}
	for _, tt := range tests {
		got := IsValidHex(tt.input)
		if got != tt.want {
			t.Errorf("IsValidHex(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestPaletteSize(t *testing.T) {
	if len(Palette) != 12 {
		t.Errorf("Palette should have 12 colors, got %d", len(Palette))
	}
}

func TestPaletteAllValidHex(t *testing.T) {
	for i, c := range Palette {
		if !IsValidHex(c) {
			t.Errorf("Palette[%d] = %q is not a valid HEX color", i, c)
		}
	}
}
