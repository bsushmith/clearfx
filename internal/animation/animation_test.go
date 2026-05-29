package animation

import "testing"

func TestRegisteredStylesProduceFrames(t *testing.T) {
	want := map[string]bool{
		"black-hole":       false,
		"fire":             false,
		"glitch":           false,
		"great-wave":       false,
		"lightning":        false,
		"matrix-rain":      false,
		"meteor-shower":    false,
		"ocean-wave":       false,
		"page-burn":        false,
		"starfield":        false,
		"typewriter-erase": false,
	}
	for _, name := range Names() {
		if _, ok := want[name]; ok {
			want[name] = true
		}
		style, err := Get(name)
		if err != nil {
			t.Fatal(err)
		}
		frame := style.New(40, 16, Options{Intensity: "medium"}).Frame(0.5)
		if frame.Width != 40 || frame.Height != 16 {
			t.Fatalf("%s frame size = %dx%d", name, frame.Width, frame.Height)
		}
		if len(frame.Cells) != 40*16 {
			t.Fatalf("%s cell count = %d", name, len(frame.Cells))
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("style %s was not registered", name)
		}
	}
}

func TestUnknownStyle(t *testing.T) {
	if _, err := Get("nope"); err == nil {
		t.Fatal("expected unknown style error")
	}
}

func TestOceanWaveChangesAcrossFrames(t *testing.T) {
	style, err := Get("ocean-wave")
	if err != nil {
		t.Fatal(err)
	}
	anim := style.New(80, 24, Options{Intensity: "medium"})
	first := anim.Frame(0.2)
	second := anim.Frame(0.7)

	if framesEqual(first, second) {
		t.Fatal("expected ocean-wave frames to change over time")
	}
}

func TestGreatWaveIncludesCurlAndFoam(t *testing.T) {
	style, err := Get("great-wave")
	if err != nil {
		t.Fatal(err)
	}
	frame := style.New(80, 24, Options{Intensity: "medium"}).Frame(0.5)

	var foam, crest int
	for _, cell := range frame.Cells {
		if cell.Color == ColorBrightWhite {
			foam++
		}
		if cell.Ch == '^' || cell.Ch == '\'' || cell.Ch == '*' {
			crest++
		}
	}
	if foam == 0 {
		t.Fatal("expected great-wave frame to include white foam")
	}
	if crest == 0 {
		t.Fatal("expected great-wave frame to include crest characters")
	}
}

func TestNewEffectsChangeAcrossFrames(t *testing.T) {
	for _, name := range []string{
		"matrix-rain",
		"glitch",
		"starfield",
		"black-hole",
		"meteor-shower",
		"lightning",
		"page-burn",
		"typewriter-erase",
	} {
		style, err := Get(name)
		if err != nil {
			t.Fatal(err)
		}
		anim := style.New(80, 24, Options{Intensity: "medium"})
		first := anim.Frame(0.15)
		second := anim.Frame(0.75)
		if framesEqual(first, second) {
			t.Fatalf("expected %s frames to change over time", name)
		}
	}
}

func framesEqual(a, b Frame) bool {
	if a.Width != b.Width || a.Height != b.Height || len(a.Cells) != len(b.Cells) {
		return false
	}
	for i := range a.Cells {
		if a.Cells[i] != b.Cells[i] {
			return false
		}
	}
	return true
}
