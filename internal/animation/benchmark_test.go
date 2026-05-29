package animation

import "testing"

func BenchmarkStylesFrame(b *testing.B) {
	const (
		width  = 120
		height = 40
	)
	opts := Options{Intensity: "medium"}
	for _, style := range List() {
		b.Run(style.Name(), func(b *testing.B) {
			anim := style.New(width, height, opts)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = anim.Frame(float64(i%60) / 59)
			}
		})
	}
}
