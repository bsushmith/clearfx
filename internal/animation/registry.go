package animation

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type Options struct {
	Duration  time.Duration
	FPS       int
	Intensity string
	Palette   string
}

type Style interface {
	Name() string
	Description() string
	New(width, height int, opts Options) Animator
}

type Animator interface {
	Frame(t float64) Frame
}

type Metadata struct {
	Name                string
	Description         string
	Category            string
	RecommendedDuration time.Duration
	RecommendedFPS      int
	SupportedPalettes   []string
	Experimental        bool
}

var styles = map[string]Style{}

func Register(style Style) {
	styles[style.Name()] = style
}

func Names() []string {
	names := make([]string, 0, len(styles))
	for name := range styles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func List() []Style {
	names := Names()
	out := make([]Style, 0, len(names))
	for _, name := range names {
		out = append(out, styles[name])
	}
	return out
}

func ListMetadata() []Metadata {
	names := Names()
	out := make([]Metadata, 0, len(names))
	for _, name := range names {
		out = append(out, MetadataFor(name))
	}
	return out
}

func Get(name string) (Style, error) {
	style, ok := styles[name]
	if !ok {
		return nil, fmt.Errorf("unknown style %q; available styles: %s", name, strings.Join(Names(), ", "))
	}
	return style, nil
}

func RandomName() string {
	names := Names()
	if len(names) == 0 {
		return ""
	}
	return names[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(names))]
}

func MetadataFor(name string) Metadata {
	style, ok := styles[name]
	if !ok {
		return Metadata{Name: name}
	}
	meta := Metadata{
		Name:                style.Name(),
		Description:         style.Description(),
		Category:            "effect",
		RecommendedDuration: 700 * time.Millisecond,
		RecommendedFPS:      30,
		SupportedPalettes:   PaletteNames(),
	}
	if override, ok := styleMetadata[name]; ok {
		if override.Category != "" {
			meta.Category = override.Category
		}
		if override.RecommendedDuration > 0 {
			meta.RecommendedDuration = override.RecommendedDuration
		}
		if override.RecommendedFPS > 0 {
			meta.RecommendedFPS = override.RecommendedFPS
		}
		if len(override.SupportedPalettes) > 0 {
			meta.SupportedPalettes = append([]string(nil), override.SupportedPalettes...)
		}
		meta.Experimental = override.Experimental
	}
	return meta
}

var styleMetadata = map[string]Metadata{
	"aurora":           {Category: "ambient", RecommendedDuration: 1200 * time.Millisecond, RecommendedFPS: 45, Experimental: true},
	"black-hole":       {Category: "collapse", RecommendedDuration: 900 * time.Millisecond, RecommendedFPS: 45},
	"dune-worm":        {Category: "creature", RecommendedDuration: 1200 * time.Millisecond, RecommendedFPS: 45, Experimental: true},
	"fire":             {Category: "elemental", RecommendedDuration: 700 * time.Millisecond, RecommendedFPS: 30},
	"glitch":           {Category: "digital", RecommendedDuration: 550 * time.Millisecond, RecommendedFPS: 45},
	"great-wave":       {Category: "water", RecommendedDuration: 1200 * time.Millisecond, RecommendedFPS: 45},
	"ink-drop":         {Category: "wipe", RecommendedDuration: 850 * time.Millisecond, RecommendedFPS: 30, Experimental: true},
	"laser-grid":       {Category: "scan", RecommendedDuration: 650 * time.Millisecond, RecommendedFPS: 45, Experimental: true},
	"lightning":        {Category: "weather", RecommendedDuration: 450 * time.Millisecond, RecommendedFPS: 45},
	"matrix-rain":      {Category: "digital", RecommendedDuration: 900 * time.Millisecond, RecommendedFPS: 45},
	"meteor-shower":    {Category: "space", RecommendedDuration: 800 * time.Millisecond, RecommendedFPS: 45},
	"ocean-wave":       {Category: "water", RecommendedDuration: 1000 * time.Millisecond, RecommendedFPS: 45},
	"page-burn":        {Category: "wipe", RecommendedDuration: 900 * time.Millisecond, RecommendedFPS: 30},
	"rainstorm":        {Category: "weather", RecommendedDuration: 900 * time.Millisecond, RecommendedFPS: 45, Experimental: true},
	"sandstorm":        {Category: "weather", RecommendedDuration: 850 * time.Millisecond, RecommendedFPS: 45, Experimental: true},
	"shatter":          {Category: "destruction", RecommendedDuration: 850 * time.Millisecond, RecommendedFPS: 45, Experimental: true},
	"snowfall":         {Category: "weather", RecommendedDuration: 1200 * time.Millisecond, RecommendedFPS: 30, Experimental: true},
	"starfield":        {Category: "space", RecommendedDuration: 850 * time.Millisecond, RecommendedFPS: 45},
	"typewriter-erase": {Category: "text", RecommendedDuration: 1000 * time.Millisecond, RecommendedFPS: 45},
}

func clamp01(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}
