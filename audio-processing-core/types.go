package audioprocessing

const DefaultSampleRate = 48_000

// EmphasisMode selects the first-order FM emphasis network.
type EmphasisMode uint8

const (
	EmphasisOff EmphasisMode = iota
	PreEmphasis
	DeEmphasis
)

// NoiseSpeed controls how quickly the reducer closes after wanted audio ends.
type NoiseSpeed string

const (
	NoiseSlow   NoiseSpeed = "slow"
	NoiseNormal NoiseSpeed = "normal"
	NoiseFast   NoiseSpeed = "fast"
)

type NoiseStatus struct {
	ReductionDB float32
	Signal      float32
	Learning    bool
}

type NotchConfig struct {
	Enabled     bool
	FrequencyHz float64
	WidthHz     float64
	DepthDB     float64
}

type CompressorConfig struct {
	Enabled    bool
	TargetDBFS float64
}

type CompressorStatus struct {
	GainDB float32
}

type Config struct {
	LowCutHz, HighCutHz  float64
	EQEnabled            bool
	EQGainsDB            [5]float64
	Notch                NotchConfig
	NoiseEnabled         bool
	NoiseBlankerEnabled  bool
	NoiseStrength        float64
	NoiseSpeed           NoiseSpeed
	Emphasis             EmphasisMode
	EmphasisMicroseconds int
	SSBClarityDB         float64
	OutputGainDB         float64
	ProcessedMix         float64
	Compressor           CompressorConfig
}

func DefaultConfig() Config {
	return Config{
		LowCutHz: 100, HighCutHz: 4000, EQEnabled: true,
		Notch:         NotchConfig{FrequencyHz: 1000, WidthHz: 120, DepthDB: -35},
		NoiseStrength: 50, NoiseSpeed: NoiseNormal,
		EmphasisMicroseconds: 50,
		SSBClarityDB:         0,
		OutputGainDB:         0,
		ProcessedMix:         100,
		Compressor: CompressorConfig{
			TargetDBFS: -18,
		},
	}
}
