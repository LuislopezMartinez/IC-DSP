package audioprocessing_test

import (
	"github.com/LuislopezMartinez/IC-DSP/audio-processing-core"
	"fmt"
)

func ExampleProcessor() {
	p := audioprocessing.NewProcessor(48_000)
	cfg := audioprocessing.DefaultConfig()
	cfg.NoiseEnabled = true
	cfg.NoiseStrength = 50
	cfg.EQGainsDB = [5]float64{2, 1, 0, -1, -2}
	cfg.Notch = audioprocessing.NotchConfig{Enabled: true, FrequencyHz: 1000, WidthHz: 120, DepthDB: -35}
	cfg.Emphasis = audioprocessing.DeEmphasis
	cfg.EmphasisMicroseconds = 50
	p.Configure(cfg)

	block := make([]float32, 480) // 10 ms of mono PCM supplied by the host program
	p.Process(block)
	fmt.Println(len(block))
	// Output: 480
}
