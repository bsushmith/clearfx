package animation

import "sort"

type Palette struct {
	Name        string
	Description string
	Primary     ANSIColor
	Secondary   ANSIColor
	Accent      ANSIColor
	Highlight   ANSIColor
	Dim         ANSIColor
	Warm        ANSIColor
	Cool        ANSIColor
	Alert       ANSIColor
	Neutral     ANSIColor
}

var palettes = map[string]Palette{
	"classic": {
		Name:        "classic",
		Description: "balanced terminal colors with a neutral white/blue base",
		Primary:     ColorBrightWhite,
		Secondary:   ColorBrightCyan,
		Accent:      ColorBrightYellow,
		Highlight:   ColorBrightRed,
		Dim:         ColorBrightBlack,
		Warm:        ColorYellow,
		Cool:        ColorCyan,
		Alert:       ColorBrightMagenta,
		Neutral:     ColorWhite,
	},
	"ember": {
		Name:        "ember",
		Description: "warm reds and oranges with bright flare accents",
		Primary:     ColorBrightWhite,
		Secondary:   ColorBrightYellow,
		Accent:      ColorBrightRed,
		Highlight:   ColorBrightMagenta,
		Dim:         ColorBrightBlack,
		Warm:        ColorRed,
		Cool:        ColorBrightCyan,
		Alert:       ColorBrightYellow,
		Neutral:     ColorWhite,
	},
	"ocean": {
		Name:        "ocean",
		Description: "deep blues and bright cyan foam",
		Primary:     ColorBrightWhite,
		Secondary:   ColorBrightCyan,
		Accent:      ColorBlue,
		Highlight:   ColorBrightBlue,
		Dim:         ColorBrightBlack,
		Warm:        ColorBrightYellow,
		Cool:        ColorCyan,
		Alert:       ColorBrightMagenta,
		Neutral:     ColorWhite,
	},
	"matrix": {
		Name:        "matrix",
		Description: "green monochrome code-rain styling",
		Primary:     ColorBrightGreen,
		Secondary:   ColorGreen,
		Accent:      ColorBrightWhite,
		Highlight:   ColorBrightGreen,
		Dim:         ColorBrightBlack,
		Warm:        ColorBrightYellow,
		Cool:        ColorBrightCyan,
		Alert:       ColorBrightWhite,
		Neutral:     ColorWhite,
	},
	"aurora": {
		Name:        "aurora",
		Description: "northern-lights tones with cyan, green, and magenta",
		Primary:     ColorBrightGreen,
		Secondary:   ColorBrightCyan,
		Accent:      ColorBrightMagenta,
		Highlight:   ColorBrightWhite,
		Dim:         ColorBrightBlack,
		Warm:        ColorBrightYellow,
		Cool:        ColorCyan,
		Alert:       ColorBrightMagenta,
		Neutral:     ColorWhite,
	},
	"monochrome": {
		Name:        "monochrome",
		Description: "white, gray, and black for subtle output",
		Primary:     ColorBrightWhite,
		Secondary:   ColorWhite,
		Accent:      ColorBrightBlack,
		Highlight:   ColorBrightWhite,
		Dim:         ColorBrightBlack,
		Warm:        ColorWhite,
		Cool:        ColorWhite,
		Alert:       ColorBrightWhite,
		Neutral:     ColorWhite,
	},
	"solarized": {
		Name:        "solarized",
		Description: "muted blue, cyan, and amber inspired by solarized palettes",
		Primary:     ColorBrightCyan,
		Secondary:   ColorCyan,
		Accent:      ColorBrightYellow,
		Highlight:   ColorBrightWhite,
		Dim:         ColorBrightBlack,
		Warm:        ColorYellow,
		Cool:        ColorBrightBlue,
		Alert:       ColorBrightMagenta,
		Neutral:     ColorWhite,
	},
}

func PaletteNames() []string {
	names := make([]string, 0, len(palettes))
	for name := range palettes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func PaletteFor(name string) Palette {
	if palette, ok := palettes[name]; ok {
		return palette
	}
	return palettes["classic"]
}
