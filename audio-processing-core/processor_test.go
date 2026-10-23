package audioprocessing

import (
	"math"
	"testing"
)

func tone(rate, n int, hz, level float64) []float32 {
	s := make([]float32, n)
	for i := range s {
		s[i] = float32(level * math.Sin(2*math.Pi*hz*float64(i)/float64(rate)))
	}
	return s
}

func rms(s []float32) float64 {
	v := 0.0
	for _, x := range s {
		v += float64(x * x)
	}
	return math.Sqrt(v / float64(len(s)))
}

func TestProcessorFiniteAndBounded(t *testing.T) {
	p := NewProcessor(48000)
	c := DefaultConfig()
	c.NoiseEnabled = true
	c.EQGainsDB = [5]float64{3, -2, 1.5, 0, -1}
	p.Configure(c)
	s := tone(48000, 48000, 1000, .8)
	p.Process(s)
	for i, v := range s {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) || v < -.98 || v > .98 {
			t.Fatalf("invalid sample %d: %v", i, v)
		}
	}
}

func TestNotch(t *testing.T) {
	measure := func(enabled bool) float64 {
		p := NewProcessor(48000)
		c := DefaultConfig()
		c.LowCutHz, c.HighCutHz, c.EQEnabled = 20, 16000, false
		c.Notch.Enabled, c.Notch.FrequencyHz, c.Notch.DepthDB = enabled, 1000, -40
		p.Configure(c)
		s := tone(48000, 48000, 1000, .4)
		p.Process(s)
		return rms(s[24000:])
	}
	if ratio := measure(true) / measure(false); ratio > .05 {
		t.Fatalf("notch ratio %.3f", ratio)
	}
}

func TestSpectrumAndAutoNotch(t *testing.T) {
	p := NewProcessor(48000)
	c := DefaultConfig()
	c.LowCutHz, c.HighCutHz, c.EQEnabled = 20, 16000, false
	p.Configure(c)
	s := tone(48000, 4096, 1000, .5)
	p.Process(s)
	db := make([]float32, 129)
	p.SpectrumBeforeNotch(db, 16000)
	hz, ok := DetectTone(db, 16000, 200, 6000)
	if !ok || math.Abs(hz-1000) > 150 {
		t.Fatalf("tone %.0f ok=%v", hz, ok)
	}
}

func TestDisabledNoiseDoesNotAddStartupSilence(t *testing.T) {
	p := NewProcessor(48000)
	c := DefaultConfig()
	c.LowCutHz, c.HighCutHz, c.EQEnabled, c.NoiseEnabled = 20, 16000, false, false
	p.Configure(c)
	s := tone(48000, 1024, 1000, .5)
	p.Process(s)
	if got := rms(s[128:]); got < .1 {
		t.Fatalf("disabled noise reducer unexpectedly muted/delayed audio: rms %.4f", got)
	}
}

func TestOutputGain(t *testing.T) {
	measure := func(gainDB float64) float64 {
		p := NewProcessor(48000)
		c := DefaultConfig()
		c.LowCutHz, c.HighCutHz = 20, 16000
		c.EQEnabled, c.NoiseEnabled = false, false
		c.OutputGainDB = gainDB
		p.Configure(c)
		s := tone(48000, 4096, 1000, .2)
		p.Process(s)
		return rms(s[512:])
	}
	ratio := measure(-6) / measure(0)
	if math.Abs(ratio-math.Pow(10, -6.0/20)) > .02 {
		t.Fatalf("unexpected -6 dB output gain ratio: %.3f", ratio)
	}
}

func TestAutomaticLeveler(t *testing.T) {
	measure := func(amplitude float64) (float64, float32) {
		p := NewProcessor(48000)
		c := DefaultConfig()
		c.LowCutHz, c.HighCutHz = 20, 16000
		c.EQEnabled, c.NoiseEnabled = false, false
		c.Compressor = CompressorConfig{Enabled: true, TargetDBFS: -18}
		p.Configure(c)
		s := tone(48000, 96000, 1000, amplitude)
		p.Process(s)
		return rms(s[72000:]), p.CompressorStatus().GainDB
	}
	quietRMS, quietGain := measure(.01)
	if quietRMS <= .025 || quietGain <= 6 {
		t.Fatalf("leveler did not amplify quiet audio: rms %.3f, gain %.2f dB", quietRMS, quietGain)
	}
	loudRMS, loudGain := measure(.8)
	if loudRMS >= .25 || loudGain >= -6 {
		t.Fatalf("leveler did not attenuate loud audio: rms %.3f, gain %.2f dB", loudRMS, loudGain)
	}
}

func TestResponseCurveIncludesNotchAndEQ(t *testing.T) {
	c := DefaultConfig()
	c.LowCutHz, c.HighCutHz = 20, 16000
	c.Notch = NotchConfig{Enabled: true, FrequencyHz: 1000, WidthHz: 120, DepthDB: -35}
	if atNotch := ResponseDB(c, 48000, 1000); atNotch > -25 {
		t.Fatalf("notch is missing from response curve: %.1f dB", atNotch)
	}
	c.Notch.Enabled = false
	c.EQGainsDB[3] = 9
	if atEQ := ResponseDB(c, 48000, 2000); atEQ < 6 {
		t.Fatalf("EQ boost is missing from response curve: %.1f dB", atEQ)
	}
}

func TestNoiseBlankerSuppressesImpulse(t *testing.T) {
	measure := func(enabled bool) float32 {
		p := NewProcessor(48000)
		c := DefaultConfig()
		c.LowCutHz, c.HighCutHz = 20, 16000
		c.EQEnabled, c.NoiseEnabled = false, false
		c.NoiseBlankerEnabled = enabled
		p.Configure(c)
		s := tone(48000, 4096, 1000, .04)
		s[2048] = .95
		p.Process(s)
		peak := float32(0)
		for _, v := range s[2048:2072] {
			if v < 0 {
				v = -v
			}
			peak = max(peak, v)
		}
		return peak
	}
	without, with := measure(false), measure(true)
	if with >= without*.55 {
		t.Fatalf("noise blanker did not sufficiently suppress impulse: off %.3f on %.3f", without, with)
	}
}

func TestSSBClarityTiltsSpeechBand(t *testing.T) {
	c := DefaultConfig()
	c.LowCutHz, c.HighCutHz = 20, 16000
	c.Emphasis = EmphasisOff
	c.SSBClarityDB = 6
	low := ResponseDB(c, 48000, 300)
	high := ResponseDB(c, 48000, 2800)
	if high-low < 4 {
		t.Fatalf("SSB clarity did not create enough speech-band tilt: low %.1f high %.1f", low, high)
	}
}

func TestDryMixReturnsUnprocessedInput(t *testing.T) {
	p := NewProcessor(48000)
	c := DefaultConfig()
	c.ProcessedMix = 0
	c.OutputGainDB = 0
	p.Configure(c)
	original := tone(48000, 4096, 1200, .1)
	processed := append([]float32(nil), original...)
	p.Process(processed)
	for i := range original {
		if math.Abs(float64(processed[i]-original[i])) > 1e-6 {
			t.Fatalf("dry crossfade changed sample %d: want %.6f got %.6f", i, original[i], processed[i])
		}
	}
}
