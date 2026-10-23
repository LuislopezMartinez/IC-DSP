package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	audio "github.com/LuislopezMartinez/IC-DSP/audio-processing-core"
	"github.com/LuislopezMartinez/IC-DSP/audio-processing-core/audioio"
	ui "github.com/LuislopezMartinez/IC-DSP/audio-processing-core/simpleui"
	"github.com/LuislopezMartinez/IC-DSP/audio-processing-core/systemtheme"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	canvasW       = 1600
	canvasH       = 1040
	maxSpectrumHz = 16000
)

type palette struct{ background, panel, panelAlt, border, text, muted, violet, cyan, blue, green, amber, red rl.Color }

var colors = palette{
	rl.Color{R: 11, G: 13, B: 24, A: 255}, rl.Color{R: 21, G: 26, B: 46, A: 255},
	rl.Color{R: 27, G: 34, B: 59, A: 255}, rl.Color{R: 56, G: 68, B: 106, A: 255},
	rl.Color{R: 244, G: 241, B: 236, A: 255}, rl.Color{R: 183, G: 189, B: 208, A: 255},
	rl.Color{R: 124, G: 92, B: 252, A: 255}, rl.Color{R: 35, G: 214, B: 217, A: 255},
	rl.Color{R: 114, G: 135, B: 245, A: 255}, rl.Color{R: 77, G: 219, B: 138, A: 255},
	rl.Color{R: 231, G: 169, B: 59, A: 255}, rl.Color{R: 232, G: 93, B: 117, A: 255},
}

func mixColor(a, b rl.Color, amount float32) rl.Color {
	mix := func(x, y uint8) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*amount) }
	return rl.Color{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B), A: 255}
}

func (a *application) applyTheme(mode string) {
	system := systemtheme.Read()
	light := mode == "light" || (mode != "dark" && system.Light)
	accent := rl.Color{R: system.Red, G: system.Green, B: system.Blue, A: 255}
	brightness := int(accent.R) + int(accent.G) + int(accent.B)
	if brightness < 150 || brightness > 700 {
		accent = rl.Color{R: 124, G: 92, B: 252, A: 255}
	}
	if light {
		colors = palette{
			background: rl.Color{R: 232, G: 235, B: 242, A: 255}, panel: rl.Color{R: 250, G: 251, B: 253, A: 255},
			panelAlt: rl.Color{R: 224, G: 228, B: 237, A: 255}, border: rl.Color{R: 151, G: 160, B: 181, A: 255},
			text: rl.Color{R: 24, G: 29, B: 42, A: 255}, muted: rl.Color{R: 76, G: 84, B: 104, A: 255}, violet: accent,
			cyan: rl.Color{R: 0, G: 145, B: 165, A: 255}, blue: rl.Color{R: 61, G: 91, B: 205, A: 255},
			green: rl.Color{R: 20, G: 148, B: 86, A: 255}, amber: rl.Color{R: 190, G: 120, B: 15, A: 255}, red: rl.Color{R: 205, G: 48, B: 72, A: 255},
		}
	} else {
		colors = palette{
			background: rl.Color{R: 18, G: 21, B: 34, A: 255}, panel: rl.Color{R: 31, G: 37, B: 57, A: 255},
			panelAlt: rl.Color{R: 40, G: 47, B: 70, A: 255}, border: rl.Color{R: 76, G: 88, B: 124, A: 255},
			text: rl.Color{R: 245, G: 246, B: 250, A: 255}, muted: rl.Color{R: 185, G: 191, B: 209, A: 255}, violet: accent,
			cyan: rl.Color{R: 35, G: 214, B: 217, A: 255}, blue: rl.Color{R: 114, G: 135, B: 245, A: 255},
			green: rl.Color{R: 77, G: 219, B: 138, A: 255}, amber: rl.Color{R: 231, G: 169, B: 59, A: 255}, red: rl.Color{R: 232, G: 93, B: 117, A: 255},
		}
	}
	ui.SetColors(colors.background, colors.background)
	theme := ui.CurrentTheme()
	theme.Text, theme.TextMuted, theme.Accent = colors.text, colors.muted, colors.violet
	theme.Control = colors.panelAlt
	theme.ControlHover = mixColor(colors.panelAlt, colors.violet, .18)
	theme.ControlPressed = colors.violet
	theme.ControlDisabled = mixColor(colors.panelAlt, colors.muted, .22)
	theme.Border, theme.SliderSelection, theme.SwitchOn = colors.border, colors.violet, colors.violet
	theme.SliderTrack = mixColor(colors.panelAlt, colors.muted, .35)
	theme.SliderHandle = mixColor(colors.text, colors.panel, .12)
	theme.SwitchOff = mixColor(colors.panelAlt, colors.muted, .42)
	theme.IndicatorOff = mixColor(colors.panelAlt, colors.muted, .5)
	theme.DisabledTrack = mixColor(colors.panelAlt, colors.text, .18)
	theme.DisabledHandle = mixColor(colors.panelAlt, colors.muted, .5)
	theme.DisabledPattern = rl.Color{R: colors.text.R, G: colors.text.G, B: colors.text.B, A: 150}
	theme.InputBackground, theme.PopupBackground, theme.PopupHover = colors.background, colors.panel, colors.panelAlt
	theme.CornerRadius = .18
	ui.SetTheme(theme)
}

type processGate struct {
	processor *audio.Processor
	active    atomic.Bool
}

type spectrumDrag uint8

const (
	dragNone spectrumDrag = iota
	dragLowCut
	dragHighCut
	dragNotchCenter
	dragNotchLeft
	dragNotchRight
)

func (g *processGate) Process(samples []float32) {
	if g.active.Load() {
		g.processor.Process(samples)
	}
}

type application struct {
	mu        sync.Mutex
	processor *audio.Processor
	gate      *processGate
	engine    *audioio.Engine
	config    audio.Config

	input, output                                                     *ui.Dropdown
	themeSelect                                                       *ui.Dropdown
	master, notchOn, noiseOn, blankerOn, eqOn, compressorOn           *ui.Switch
	lowCut, highCut, notchFreq, notchWidth, notchDepth, noiseStrength *ui.Slider
	outputGain                                                        *ui.Slider
	processedMix                                                      *ui.Slider
	compressorTarget                                                  *ui.Slider
	ssbClarity                                                        *ui.Slider
	eq                                                                [5]*ui.Slider
	noiseSpeed                                                        [3]*ui.Button
	emphasisMode                                                      [3]*ui.Button

	before, after, rawBefore, rawAfter  []float32
	graphDrag                           spectrumDrag
	lastClick                           float64
	inputClipUntil, outputClipUntil     float64
	spectrumPaused                      bool
	spectrumPauseButton                 *ui.Button
	inputPeakHold, outputPeakHold       float32
	lastMeterTime, lastReconnectAttempt float64
	preferredInput, preferredOutput     string
	themeMode                           string
	presetNames                         [3]*ui.TextField
	presetLoad                          [3]*ui.Button
	presetSave                          [3]*ui.Button
	presetSlots                         [3]presetSlot
	activeSlot                          int
	activePresetConfig                  audio.Config
	presetDirty                         bool
	pendingOverwrite                    int
	overwriteUntil                      float64
	message                             string
}

type appSettings struct {
	InputDevice  string `json:"input_device"`
	OutputDevice string `json:"output_device"`
	ActivePreset int    `json:"active_preset"`
	Theme        string `json:"theme"`
}

type presetSlot struct {
	Name   string       `json:"name"`
	Saved  bool         `json:"saved"`
	Config audio.Config `json:"config"`
}

type presetStore struct {
	Version int           `json:"version"`
	Slots   [3]presetSlot `json:"slots"`
}

func main() {
	processor := audio.NewProcessor(audio.DefaultSampleRate)
	gate := &processGate{processor: processor}
	gate.active.Store(true)
	engine, err := audioio.New(gate)
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	app := &application{processor: processor, gate: gate, engine: engine, config: audio.DefaultConfig(), before: make([]float32, 257), after: make([]float32, 257), rawBefore: make([]float32, 257), rawAfter: make([]float32, 257), inputPeakHold: -80, outputPeakHold: -80, activeSlot: -1, pendingOverwrite: -1, themeMode: "auto"}
	app.loadSettings()
	app.loadPreset()
	app.processor.Configure(app.config)
	app.buildUI()
	app.startSelectedDevices()

	ui.SetMode(canvasW, canvasH, ui.Fit)
	ui.SetInitialWindowSize(1440, 936)
	ui.SetMinimumSize(960, 624)
	ui.SetTitle("IC-DSP · Procesador de audio")
	ui.SetTargetFPS(60)
	ui.SetTextScale(1.14)
	app.applyTheme(app.themeMode)
	app.updateControlStates(app.config)
	ui.Run(app.draw)
}

func names(devices []audioio.Device) []string {
	if len(devices) == 0 {
		return []string{"Sin dispositivos"}
	}
	result := make([]string, len(devices))
	for i, d := range devices {
		result[i] = d.Name
		if d.IsDefault {
			result[i] += " · Predeterminado"
		}
	}
	return result
}
func defaultIndex(devices []audioio.Device) int {
	for i, d := range devices {
		if d.IsDefault {
			return i
		}
	}
	return 0
}

func (a *application) buildUI() {
	inputs, outputs := a.engine.Inputs(), a.engine.Outputs()
	a.themeSelect = ui.NewDropdown("theme", 1320, 15, 240, 34, "Tema", []string{"Automático", "Claro", "Oscuro"}, 13)
	themeIndex := 0
	if a.themeMode == "light" {
		themeIndex = 1
	} else if a.themeMode == "dark" {
		themeIndex = 2
	}
	a.themeSelect.SetSelected(themeIndex)
	a.themeSelect.OnChange(func(index int, _ string) {
		a.themeMode = []string{"auto", "light", "dark"}[index]
		a.applyTheme(a.themeMode)
		a.updateControlStates(a.configSnapshot())
		a.saveSettings()
	})
	ui.Add(a.themeSelect)
	a.input = ui.NewDropdown("input", 425, 70, 290, 46, "Dispositivo de entrada", names(inputs), 16)
	a.output = ui.NewDropdown("output", 950, 70, 280, 46, "Dispositivo de salida", names(outputs), 16)
	a.input.SetSelected(selectedDeviceIndex(inputs, a.preferredInput))
	a.output.SetSelected(selectedDeviceIndex(outputs, a.preferredOutput))
	a.input.OnChange(func(int, string) { a.startSelectedDevices() })
	a.output.OnChange(func(int, string) { a.startSelectedDevices() })
	ui.Add(a.input)
	ui.Add(a.output)
	a.processedMix = a.slider("processed-mix", 748, 77, 130, 28, 0, 100, float32(a.config.ProcessedMix), 1, func(v float32) {
		a.update(func(c *audio.Config) { c.ProcessedMix = float64(v) })
	})
	refresh := ui.NewButton("refresh-devices", 1237, 78, 28, 28, "↻", 17)
	refresh.OnClick(a.refreshDevices)
	ui.Add(refresh)

	a.master = ui.NewSwitch("master", 1270, 76, 285, 30, "PROCESAMIENTO ACTIVO", true, 15)
	a.master.OnChange(func(on bool) { a.gate.active.Store(on) })
	ui.Add(a.master)
	a.spectrumPauseButton = ui.NewButton("spectrum-pause", 565, 151, 92, 27, "PAUSAR", 11)
	a.spectrumPauseButton.OnClick(func() {
		a.spectrumPaused = !a.spectrumPaused
		if a.spectrumPaused {
			a.spectrumPauseButton.SetLabel("REANUDAR")
		} else {
			a.spectrumPauseButton.SetLabel("PAUSAR")
		}
	})
	ui.Add(a.spectrumPauseButton)

	a.lowCut = a.slider("low", 1340, 178, 135, 28, 20, 4000, float32(a.config.LowCutHz), 10, func(v float32) { a.update(func(c *audio.Config) { c.LowCutHz = float64(v) }) })
	a.highCut = a.slider("high", 1340, 218, 135, 28, 250, 16000, float32(a.config.HighCutHz), 50, func(v float32) { a.update(func(c *audio.Config) { c.HighCutHz = float64(v) }) })
	a.notchOn = ui.NewSwitch("notch-on", 1506, 278, 54, 30, "", a.config.Notch.Enabled, 14)
	a.notchOn.OnChange(func(on bool) { a.update(func(c *audio.Config) { c.Notch.Enabled = on }) })
	ui.Add(a.notchOn)
	a.notchFreq = a.slider("notch-freq", 1340, 314, 135, 28, 80, maxSpectrumHz, float32(a.config.Notch.FrequencyHz), 1, func(v float32) { a.update(func(c *audio.Config) { c.Notch.FrequencyHz = float64(v) }) })
	a.notchWidth = a.slider("notch-width", 1340, 350, 135, 28, 20, 2000, float32(a.config.Notch.WidthHz), 5, func(v float32) { a.update(func(c *audio.Config) { c.Notch.WidthHz = float64(v) }) })
	a.notchDepth = a.slider("notch-depth", 1340, 386, 135, 28, -60, -6, float32(a.config.Notch.DepthDB), 1, func(v float32) { a.update(func(c *audio.Config) { c.Notch.DepthDB = float64(v) }) })
	detect := ui.NewButton("detect", 1250, 428, 330, 36, "DETECTAR TONO", 14)
	detect.OnClick(a.detectTone)
	ui.Add(detect)

	a.noiseOn = ui.NewSwitch("noise-on", 1506, 493, 54, 30, "", a.config.NoiseEnabled, 14)
	a.noiseOn.OnChange(func(on bool) {
		a.update(func(c *audio.Config) { c.NoiseEnabled = on })
		if on {
			a.processor.ResetNoiseLearning()
		}
	})
	ui.Add(a.noiseOn)
	a.noiseStrength = a.slider("noise-strength", 1340, 532, 135, 28, 0, 100, float32(a.config.NoiseStrength), 1, func(v float32) { a.update(func(c *audio.Config) { c.NoiseStrength = float64(v) }) })
	for i, item := range []struct {
		label string
		speed audio.NoiseSpeed
	}{{"Lenta", audio.NoiseSlow}, {"Normal", audio.NoiseNormal}, {"Rápida", audio.NoiseFast}} {
		button := ui.NewButton("speed-"+item.label, 1240+float32(i)*112, 571, 108, 34, item.label, 13)
		speed := item.speed
		button.OnClick(func() { a.update(func(c *audio.Config) { c.NoiseSpeed = speed }) })
		a.noiseSpeed[i] = button
		ui.Add(button)
	}

	a.eqOn = ui.NewSwitch("eq-on", 1048, 643, 118, 26, "ACTIVO", a.config.EQEnabled, 13)
	a.eqOn.OnChange(func(on bool) { a.update(func(c *audio.Config) { c.EQEnabled = on }) })
	ui.Add(a.eqOn)
	eqReset := ui.NewButton("eq-reset", 920, 643, 100, 26, "RESET EQ", 10)
	eqReset.OnClick(func() {
		a.update(func(c *audio.Config) { c.EQGainsDB = [5]float64{} })
		for _, slider := range a.eq {
			if slider != nil {
				slider.SetValue(0)
			}
		}
		a.message = "Ecualizador restablecido"
	})
	ui.Add(eqReset)
	for i := range a.eq {
		x := 610 + float32(i)*110
		s := ui.NewSlider(fmt.Sprintf("eq-%d", i), x, 702, 32, 158, -12, 12, float32(a.config.EQGainsDB[i]))
		s.SetOrientation(ui.Vertical)
		s.SetStep(.5)
		band := i
		s.OnChange(func(v float32) { a.update(func(c *audio.Config) { c.EQGainsDB[band] = float64(v) }) })
		a.eq[i] = s
		ui.Add(s)
	}

	for i, item := range []struct {
		label string
		us    int
		ssb   bool
	}{{"FM 50 µs", 50, false}, {"FM 75 µs", 75, false}, {"SSB", 0, true}} {
		button := ui.NewButton("radio-mode-"+item.label, 1225+float32(i)*118, 690, 114, 35, item.label, 11)
		mode := item
		button.OnClick(func() {
			a.update(func(c *audio.Config) {
				if mode.ssb {
					c.Emphasis = audio.EmphasisOff
				} else {
					c.Emphasis = audio.DeEmphasis
					c.EmphasisMicroseconds = mode.us
				}
			})
		})
		a.emphasisMode[i] = button
		ui.Add(button)
	}
	a.ssbClarity = a.slider("ssb-clarity", 1340, 735, 135, 28, -6, 9, float32(a.config.SSBClarityDB), .5, func(v float32) {
		a.update(func(c *audio.Config) { c.SSBClarityDB = float64(v) })
	})
	a.blankerOn = ui.NewSwitch("blanker-on", 1506, 802, 54, 30, "", a.config.NoiseBlankerEnabled, 14)
	a.blankerOn.OnChange(func(on bool) { a.update(func(c *audio.Config) { c.NoiseBlankerEnabled = on }) })
	ui.Add(a.blankerOn)
	a.outputGain = a.slider("output-gain", 105, 79, 145, 28, -60, 12, float32(a.config.OutputGainDB), .5, func(v float32) {
		a.update(func(c *audio.Config) { c.OutputGainDB = float64(v) })
	})
	a.compressorOn = ui.NewSwitch("compressor-on", 235, 902, 54, 30, "", a.config.Compressor.Enabled, 14)
	a.compressorOn.OnChange(func(on bool) { a.update(func(c *audio.Config) { c.Compressor.Enabled = on }) })
	ui.Add(a.compressorOn)
	a.compressorTarget = a.slider("compressor-target", 455, 920, 455, 30, -30, -6, float32(a.config.Compressor.TargetDBFS), 1, func(v float32) {
		a.update(func(c *audio.Config) { c.Compressor.TargetDBFS = float64(v) })
	})
	reset := ui.NewButton("reset", 1435, 965, 130, 34, "Restablecer", 12)
	reset.OnClick(a.reset)
	ui.Add(reset)
	for i := 0; i < 3; i++ {
		y := 666 + float32(i)*69
		name := ui.NewTextField(fmt.Sprintf("preset-name-%d", i), 65, y, 300, 34, "Nombre del preset", 14)
		name.SetMaxLength(22)
		name.SetText(a.presetSlots[i].Name)
		a.presetNames[i] = name
		ui.Add(name)
		load := ui.NewButton(fmt.Sprintf("preset-load-%d", i), 375, y, 62, 34, "CARGAR", 10)
		slot := i
		load.OnClick(func() { a.loadPresetSlot(slot) })
		a.presetLoad[i] = load
		ui.Add(load)
		save := ui.NewButton(fmt.Sprintf("preset-save-%d", i), 445, y, 80, 34, "GUARDAR", 10)
		save.OnClick(func() { a.savePresetSlot(slot) })
		a.presetSave[i] = save
		ui.Add(save)
	}
	a.updatePresetButtons()
	a.updateControlStates(a.config)
}

func (a *application) slider(id string, x, y, w, h, lo, hi, value, step float32, change func(float32)) *ui.Slider {
	s := ui.NewSlider(id, x, y, w, h, lo, hi, value)
	s.SetStep(step)
	s.OnChange(change)
	ui.Add(s)
	return s
}

func (a *application) update(change func(*audio.Config)) {
	a.mu.Lock()
	change(&a.config)
	cfg := a.config
	a.mu.Unlock()
	a.processor.Configure(cfg)
	a.updateControlStates(cfg)
	if a.activeSlot >= 0 {
		a.presetDirty = cfg != a.activePresetConfig
	}
}

func (a *application) updateControlStates(cfg audio.Config) {
	if a.notchFreq == nil {
		return
	}
	for _, control := range []*ui.Slider{a.notchFreq, a.notchWidth, a.notchDepth} {
		control.SetEnabled(cfg.Notch.Enabled)
	}
	a.noiseStrength.SetEnabled(cfg.NoiseEnabled)
	for _, button := range a.noiseSpeed {
		button.SetEnabled(cfg.NoiseEnabled)
	}
	for _, control := range a.eq {
		control.SetEnabled(cfg.EQEnabled)
	}
	a.compressorTarget.SetEnabled(cfg.Compressor.Enabled)
	a.ssbClarity.SetEnabled(cfg.Emphasis == audio.EmphasisOff)
	a.updateSelectedButtons(cfg)
}

func setButtonSelected(button *ui.Button, selected bool) {
	if selected {
		button.SetColors(colors.violet, rl.Color{R: 166, G: 139, B: 255, A: 255}, rl.White)
	} else {
		button.ClearColors()
	}
}

func (a *application) updateSelectedButtons(cfg audio.Config) {
	for i, speed := range []audio.NoiseSpeed{audio.NoiseSlow, audio.NoiseNormal, audio.NoiseFast} {
		setButtonSelected(a.noiseSpeed[i], cfg.NoiseSpeed == speed)
	}
	mode := 0
	if cfg.Emphasis == audio.EmphasisOff {
		mode = 2
	} else if cfg.EmphasisMicroseconds == 75 {
		mode = 1
	}
	for i, button := range a.emphasisMode {
		setButtonSelected(button, i == mode)
	}
}
func (a *application) configSnapshot() audio.Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.config
}

func (a *application) startSelectedDevices() {
	if err := a.engine.Start(a.input.SelectedIndex(), a.output.SelectedIndex()); err != nil {
		a.message = err.Error()
	} else {
		a.message = ""
		a.preferredInput, a.preferredOutput = a.input.SelectedText(), a.output.SelectedText()
		a.saveSettings()
	}
}

func selectedDeviceIndex(devices []audioio.Device, previous string) int {
	labels := names(devices)
	for i, label := range labels {
		if label == previous {
			return i
		}
	}
	return defaultIndex(devices)
}

func (a *application) refreshDevices() {
	oldInput, oldOutput := a.input.SelectedText(), a.output.SelectedText()
	if err := a.engine.Refresh(); err != nil {
		a.message = "No se pudieron actualizar los dispositivos: " + err.Error()
		return
	}
	inputs, outputs := a.engine.Inputs(), a.engine.Outputs()
	a.input.SetItems(names(inputs))
	a.output.SetItems(names(outputs))
	a.input.SetSelected(selectedDeviceIndex(inputs, oldInput))
	a.output.SetSelected(selectedDeviceIndex(outputs, oldOutput))
	a.startSelectedDevices()
	if a.message == "" {
		a.message = "Dispositivos de audio actualizados"
	}
}

func (a *application) draw() {
	if a.pendingOverwrite >= 0 && rl.GetTime() > a.overwriteUntil {
		a.presetSave[a.pendingOverwrite].SetLabel("GUARDAR")
		a.pendingOverwrite = -1
	}
	if !a.spectrumPaused {
		a.processor.SpectrumBeforeNotch(a.rawBefore, maxSpectrumHz)
		a.processor.Spectrum(a.rawAfter, maxSpectrumHz)
		for i := range a.before {
			if a.before[i] == 0 {
				a.before[i], a.after[i] = a.rawBefore[i], a.rawAfter[i]
			} else {
				a.before[i] += .28 * (a.rawBefore[i] - a.before[i])
				a.after[i] += .28 * (a.rawAfter[i] - a.after[i])
			}
		}
	}
	a.handleSpectrumInput()
	a.drawChrome()
	status := a.engine.Status()
	if !status.Running && status.Error != "" && rl.GetTime()-a.lastReconnectAttempt > 5 {
		a.lastReconnectAttempt = rl.GetTime()
		a.refreshDevices()
	}
}

func panel(x, y, w, h float32) {
	r := rl.Rectangle{X: x, Y: y, Width: w, Height: h}
	rl.DrawRectangleRounded(r, .04, 8, colors.panel)
	rl.DrawRectangleRoundedLinesEx(r, .04, 8, 1, colors.border)
}
func text(value string, x, y float32, size int32, color rl.Color) {
	ui.DrawText(value, x, y, size, color)
}
func mono(value string, x, y float32, size int32, color rl.Color) {
	ui.DrawTextStyled(value, x, y, size, ui.FontMono, color)
}
func heading(value string, x, y float32) {
	ui.DrawTextStyled(value, x, y, 15, ui.FontSemiBold, colors.text)
}

func sectionHeading(value string, x, y float32) {
	size := ui.MeasureTextStyled(value, 15, ui.FontSemiBold)
	r := rl.Rectangle{X: x - 8, Y: y - 5, Width: size.X + 16, Height: 25}
	rl.DrawRectangleRounded(r, .24, 6, colors.panelAlt)
	rl.DrawRectangleRoundedLinesEx(r, .24, 6, 1.2, colors.border)
	ui.DrawTextStyled(value, x, y, 15, ui.FontSemiBold, colors.text)
}

func (a *application) drawChrome() {
	status := a.engine.Status()
	cfg := a.configSnapshot()
	text("IC-DSP", 28, 18, 32, colors.text)
	badge := rl.Rectangle{X: 198, Y: 17, Width: 310, Height: 34}
	rl.DrawRectangleRounded(badge, .45, 12, rl.Color{R: 18, G: 59, B: 50, A: 255})
	rl.DrawCircle(218, 34, 7, colors.green)
	state := "EN ESPERA"
	if status.Running {
		state = fmt.Sprintf("EN EJECUCIÓN · %d kHz · %.1f ms", status.SampleRate/1000, float64(status.BufferFrames)*1000/float64(max(status.SampleRate, 1)))
	}
	text(state, 237, 25, 14, colors.green)
	if status.Error != "" {
		text(shorten(status.Error, 72), 530, 25, 12, colors.amber)
	} else if a.message != "" {
		text(shorten(a.message, 72), 530, 25, 12, colors.amber)
	}
	text(fmt.Sprintf("DSP %.1f%%", status.CPULoad), 1100, 25, 12, colors.muted)
	text("TEMA", 1260, 25, 12, colors.muted)
	panel(16, 58, 1568, 70)
	heading("VOLUMEN", 32, 84)
	valueBox(fmt.Sprintf("%+.1f dB", cfg.OutputGainDB), 260, 77)
	heading("ENTRADA", 350, 84)
	text("MEZCLA", 748, 63, 10, colors.muted)
	mono(fmt.Sprintf("%.0f%%", cfg.ProcessedMix), 838, 63, 10, colors.text)
	text("DIRECTA", 724, 108, 9, colors.muted)
	text("PROCESADA", 828, 108, 9, colors.muted)
	heading("SALIDA", 895, 84)

	panel(16, 140, 190, 485)
	a.drawMeters(status)
	panel(218, 140, 980, 485)
	a.drawSpectrum(cfg)
	panel(1210, 140, 374, 120)
	sectionHeading("FILTRO DE AUDIO", 1230, 158)
	text("Corte bajo", 1230, 184, 13, colors.muted)
	valueBox(fmt.Sprintf("%.0f Hz", cfg.LowCutHz), 1485, 176)
	text("Corte alto", 1230, 224, 13, colors.muted)
	valueBox(formatHz(cfg.HighCutHz), 1485, 216)
	panel(1210, 270, 374, 205)
	sectionHeading("NOTCH", 1230, 288)
	if !cfg.Notch.Enabled {
		disabledBadge("DESACTIVADO", 1360, 284)
	}
	text("Frecuencia", 1230, 318, 13, colors.muted)
	valueBox(formatHz(cfg.Notch.FrequencyHz), 1485, 307)
	text("Ancho", 1230, 354, 13, colors.muted)
	valueBox(fmt.Sprintf("%.0f Hz", cfg.Notch.WidthHz), 1485, 343)
	text("Profundidad", 1230, 390, 13, colors.muted)
	valueBox(fmt.Sprintf("%.0f dB", cfg.Notch.DepthDB), 1485, 379)
	panel(1210, 485, 374, 140)
	sectionHeading("REDUCCIÓN DE RUIDO", 1230, 503)
	if !cfg.NoiseEnabled {
		disabledBadge("DESACTIVADA", 1410, 499)
	}
	text("Intensidad", 1230, 536, 13, colors.muted)
	valueBox(fmt.Sprintf("%.0f %%", cfg.NoiseStrength), 1485, 525)
	ns := a.processor.NoiseStatus()
	learned := "Aprendiendo…"
	if !ns.Learning {
		learned = fmt.Sprintf("Aprendido · %.1f dB", ns.ReductionDB)
	}
	text(learned, 1230, 608, 12, colors.green)

	panel(16, 635, 520, 245)
	heading("PRESETS", 38, 639)
	for i := 0; i < 3; i++ {
		y := 660 + float32(i)*69
		cardColor := colors.border
		if i == a.activeSlot {
			cardColor = colors.violet
		}
		card := rl.Rectangle{X: 26, Y: y, Width: 500, Height: 65}
		rl.DrawRectangleRounded(card, .09, 6, colors.panelAlt)
		rl.DrawRectangleRoundedLinesEx(card, .09, 6, 1.5, cardColor)
		mono(fmt.Sprintf("%d", i+1), 35, y+11, 18, cardColor)
		state := "VACÍO"
		stateColor := colors.muted
		if a.presetSlots[i].Saved {
			state, stateColor = "●", colors.green
		}
		if i == a.activeSlot {
			state, stateColor = "●", colors.violet
			if a.presetDirty {
				state, stateColor = "●", colors.amber
			}
		}
		text(state, 35, y+46, 11, stateColor)
	}
	panel(548, 635, 650, 245)
	heading("ECUALIZADOR", 568, 654)
	if !cfg.EQEnabled {
		disabledBadge("DESACTIVADO", 695, 650)
	}
	text("+12 dB", 576, 704, 12, colors.muted)
	text("0 dB", 588, 778, 12, colors.muted)
	text("-12 dB", 576, 850, 12, colors.muted)
	for i, hz := range audio.EQFrequencies {
		center := 626 + float32(i)*110
		gain := fmt.Sprintf("%+.1f dB", cfg.EQGainsDB[i])
		frequency := formatHz(hz)
		text(gain, center-ui.MeasureText(gain, 12).X*.5, 684, 12, colors.text)
		text(frequency, center-ui.MeasureText(frequency, 12).X*.5, 858, 12, colors.text)
	}
	panel(1210, 635, 374, 145)
	sectionHeading("MODO / ÉNFASIS", 1230, 654)
	text("Claridad SSB", 1230, 741, 12, colors.muted)
	valueBox(fmt.Sprintf("%+.1f dB", cfg.SSBClarityDB), 1485, 733)
	text("CÁLIDA", 1340, 766, 9, colors.muted)
	text("CLARA", 1442, 766, 9, colors.muted)
	panel(1210, 790, 374, 90)
	heading("NOISE BLANKER", 1230, 808)
	text("Supresión de ruido impulsivo", 1230, 842, 12, colors.muted)
	if !cfg.NoiseBlankerEnabled {
		disabledBadge("DESACTIVADO", 1360, 804)
	}
	panel(16, 890, 1568, 130)
	heading("COMPRESOR", 38, 908)
	if !cfg.Compressor.Enabled {
		disabledBadge("DESACTIVADO", 130, 904)
	}
	text("COMPRESIÓN", 330, 927, 12, colors.muted)
	valueBox(fmt.Sprintf("%.0f dBFS", cfg.Compressor.TargetDBFS), 930, 918)
	compression := a.processor.CompressorStatus()
	text("GANANCIA AUTOMÁTICA", 1040, 907, 11, colors.muted)
	gainColor := colors.green
	if compression.GainDB < 0 {
		gainColor = colors.amber
	}
	drawGainMeter(1040, 936, 165, compression.GainDB, gainColor)
	mono(fmt.Sprintf("%+.1f dB", compression.GainDB), 1218, 929, 13, gainColor)
}

func disabledBadge(label string, x, y float32) {
	size := ui.MeasureTextStyled(label, 10, ui.FontSemiBold)
	r := rl.Rectangle{X: x, Y: y, Width: size.X + 14, Height: 21}
	rl.DrawRectangleRounded(r, .35, 6, colors.panelAlt)
	rl.DrawRectangleRoundedLinesEx(r, .35, 6, 1, colors.border)
	ui.DrawTextStyled(label, x+7, y+5, 10, ui.FontSemiBold, colors.muted)
}

func drawGainMeter(x, y, width float32, gainDB float32, active rl.Color) {
	track := rl.Rectangle{X: x, Y: y, Width: width, Height: 10}
	rl.DrawRectangleRounded(track, .45, 8, rl.Color{R: 46, G: 54, B: 80, A: 255})
	center := x + width*.5
	rl.DrawLineEx(rl.Vector2{X: center, Y: y - 3}, rl.Vector2{X: center, Y: y + 13}, 2, colors.text)
	normalized := float32(clamp64(float64(gainDB/24), -1, 1))
	end := center + normalized*width*.5
	left := float32(math.Min(float64(center), float64(end)))
	fill := rl.Rectangle{X: left, Y: y, Width: float32(math.Abs(float64(end - center))), Height: 10}
	if fill.Width > 0 {
		rl.DrawRectangleRounded(fill, .45, 8, active)
	}
	text("−24", x, y+16, 9, colors.muted)
	text("0", center-3, y+16, 9, colors.muted)
	text("+24", x+width-18, y+16, 9, colors.muted)
}

func clamp64(value, minimum, maximum float64) float64 {
	return math.Max(minimum, math.Min(maximum, value))
}

func drawWaveIcon(x, y float32) {
	for i, h := range []float32{8, 18, 30, 22, 12, 26, 15, 7} {
		xx := x + float32(i)*9
		rl.DrawLineEx(rl.Vector2{X: xx, Y: y - h/2}, rl.Vector2{X: xx, Y: y + h/2}, 2, colors.violet)
	}
}
func valueBox(value string, x, y float32) {
	r := rl.Rectangle{X: x, Y: y, Width: 80, Height: 31}
	rl.DrawRectangleRounded(r, .18, 6, colors.background)
	rl.DrawRectangleRoundedLinesEx(r, .18, 6, 1, colors.border)
	size := ui.MeasureTextStyled(value, 13, ui.FontMono)
	mono(value, x+(80-size.X)/2, y+7, 13, colors.text)
}
func formatHz(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("%.2f kHz", v/1000)
	}
	return fmt.Sprintf("%.0f Hz", v)
}
func shorten(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func (a *application) drawMeters(status audioio.Status) {
	now := rl.GetTime()
	delta := now - a.lastMeterTime
	if a.lastMeterTime == 0 || delta > 1 {
		delta = 0
	}
	a.lastMeterTime = now
	a.inputPeakHold = max(status.InputDBFS, a.inputPeakHold-float32(delta)*12)
	a.outputPeakHold = max(status.OutputDBFS, a.outputPeakHold-float32(delta)*12)
	if status.InputDBFS >= -.5 {
		a.inputClipUntil = now + 2
	}
	if status.OutputDBFS >= -.5 {
		a.outputClipUntil = now + 2
	}
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && ui.MouseInViewport() {
		pointer := ui.MousePosition()
		if rl.CheckCollisionPointRec(pointer, rl.Rectangle{X: 44, Y: 150, Width: 45, Height: 405}) {
			a.inputClipUntil = 0
		}
		if rl.CheckCollisionPointRec(pointer, rl.Rectangle{X: 115, Y: 150, Width: 45, Height: 405}) {
			a.outputClipUntil = 0
		}
	}
	heading("IN", 62, 160)
	heading("OUT", 133, 160)
	if a.inputClipUntil > now {
		drawClipBadge(48, 174)
	}
	if a.outputClipUntil > now {
		drawClipBadge(119, 174)
	}
	drawMeter(52, 190, 28, 345, status.InputDBFS)
	drawMeter(123, 190, 28, 345, status.OutputDBFS)
	drawPeakHold(52, 190, 28, 345, a.inputPeakHold)
	drawPeakHold(123, 190, 28, 345, a.outputPeakHold)
	valueBox(fmt.Sprintf("%.1f", status.InputDBFS), 28, 553)
	valueBox(fmt.Sprintf("%.1f", status.OutputDBFS), 99, 553)
	text("dBFS", 51, 589, 11, colors.muted)
	text("dBFS", 122, 589, 11, colors.muted)
	mono(fmt.Sprintf("RMS %.1f", status.InputRMSDBFS), 25, 606, 9, colors.muted)
	mono(fmt.Sprintf("RMS %.1f", status.OutputRMSDBFS), 96, 606, 9, colors.muted)
}
func drawPeakHold(x, y, w, h, db float32) {
	ratio := max(0, min(1, (db+60)/60))
	yy := y + h - ratio*h
	color := colors.green
	if db > -12 {
		color = colors.amber
	}
	if db > -1 {
		color = colors.red
	}
	rl.DrawLineEx(rl.Vector2{X: x - 3, Y: yy}, rl.Vector2{X: x + w + 3, Y: yy}, 2, color)
}
func drawClipBadge(x, y float32) {
	r := rl.Rectangle{X: x, Y: y, Width: 36, Height: 15}
	rl.DrawRectangleRounded(r, .25, 5, colors.red)
	ui.DrawTextStyled("CLIP", x+6, y+3, 9, ui.FontSemiBold, rl.White)
}
func drawMeter(x, y, w, h, db float32) {
	rl.DrawRectangleRounded(rl.Rectangle{X: x, Y: y, Width: w, Height: h}, .18, 5, colors.background)
	segments := 18
	lit := int(float32(segments) * max(0, min(1, (db+60)/60)))
	for i := 0; i < segments; i++ {
		yy := y + h - 11 - float32(i)*18
		c := rl.Color{R: 43, G: 49, B: 70, A: 255}
		if i < lit {
			c = colors.green
			if i > 12 {
				c = colors.amber
			}
			if i > 16 {
				c = colors.red
			}
		}
		rl.DrawRectangle(int32(x+6), int32(yy), int32(w-12), 10, c)
	}
}

func (a *application) drawSpectrum(cfg audio.Config) {
	heading("ESPECTRO", 238, 160)
	legend(700, 164, colors.amber, "Respuesta")
	legend(890, 164, colors.blue, "Antes del notch")
	legend(1040, 164, colors.cyan, "Salida")
	g := rl.Rectangle{X: 280, Y: 205, Width: 880, Height: 370}
	rl.DrawRectangleRec(g, colors.background)
	lowX, highX := spectrumX(g, cfg.LowCutHz), spectrumX(g, cfg.HighCutHz)
	// Rejected regions remain visible but subdued; the accepted band receives a
	// very light tint so the filter limits read as one coherent control.
	rl.DrawRectangleRec(rl.Rectangle{X: g.X, Y: g.Y, Width: max(0, lowX-g.X), Height: g.Height}, rl.Color{R: 7, G: 8, B: 17, A: 190})
	rl.DrawRectangleRec(rl.Rectangle{X: lowX, Y: g.Y, Width: max(0, highX-lowX), Height: g.Height}, rl.Color{R: 82, G: 62, B: 170, A: 22})
	rl.DrawRectangleRec(rl.Rectangle{X: highX, Y: g.Y, Width: max(0, g.X+g.Width-highX), Height: g.Height}, rl.Color{R: 7, G: 8, B: 17, A: 190})
	for i := 0; i <= 8; i++ {
		x := g.X + g.Width*float32(i)/8
		rl.DrawLine(int32(x), int32(g.Y), int32(x), int32(g.Y+g.Height), rl.Color{R: 40, G: 48, B: 79, A: 255})
		text(fmt.Sprintf("%dk", i*2), x-8, g.Y+g.Height+10, 11, colors.muted)
	}
	for i := 0; i <= 4; i++ {
		y := g.Y + g.Height*float32(i)/4
		rl.DrawLine(int32(g.X), int32(y), int32(g.X+g.Width), int32(y), rl.Color{R: 40, G: 48, B: 79, A: 255})
		text(fmt.Sprintf("%d", -i*20), g.X-34, y-6, 11, colors.muted)
	}
	drawTrace(g, a.before, colors.blue)
	drawTrace(g, a.after, colors.cyan)
	drawResponseCurve(g, cfg)
	drawCutoffHandle(g, lowX, "CORTE BAJO · "+formatHz(cfg.LowCutHz), true, a.graphDrag == dragLowCut)
	drawCutoffHandle(g, highX, "CORTE ALTO · "+formatHz(cfg.HighCutHz), false, a.graphDrag == dragHighCut)
	drawPassbandBracket(g, lowX, highX)
	a.drawNotchOverlay(g, cfg)
}

func drawResponseCurve(g rl.Rectangle, cfg audio.Config) {
	const points = 320
	var previous rl.Vector2
	for i := 0; i < points; i++ {
		x := g.X + g.Width*float32(i)/float32(points-1)
		hz := max(20, spectrumHz(g, x))
		db := audio.ResponseDB(cfg, audio.DefaultSampleRate, hz)
		y := g.Y + float32((18-db)/48)*g.Height
		y = max(g.Y, min(g.Y+g.Height, y))
		current := rl.Vector2{X: x, Y: y}
		if i > 0 {
			rl.DrawLineEx(previous, current, 2.2, colors.amber)
		}
		previous = current
	}
	text("+18", g.X+g.Width-28, g.Y+4, 9, colors.amber)
	text("0", g.X+g.Width-12, g.Y+g.Height*.375-5, 9, colors.amber)
	text("−30", g.X+g.Width-28, g.Y+g.Height-14, 9, colors.amber)
}

func spectrumX(g rl.Rectangle, hz float64) float32 {
	return g.X + float32(max(0, min(maxSpectrumHz, hz))/maxSpectrumHz)*g.Width
}

func spectrumHz(g rl.Rectangle, x float32) float64 {
	return max(0, min(maxSpectrumHz, float64((x-g.X)/g.Width*maxSpectrumHz)))
}

func drawCutoffHandle(g rl.Rectangle, x float32, label string, left, active bool) {
	c := colors.violet
	if active {
		c = rl.Color{R: 166, G: 139, B: 255, A: 255}
	}
	rl.DrawLineEx(rl.Vector2{X: x, Y: g.Y}, rl.Vector2{X: x, Y: g.Y + g.Height}, 2.5, c)
	drawGraphHandle(x, g.Y+10, c, active)
	sz := ui.MeasureTextStyled(label, 11, ui.FontSemiBold)
	boxX := x + 10
	if !left {
		boxX = x - sz.X - 26
	}
	boxX = max(g.X, min(g.X+g.Width-sz.X-16, boxX))
	box := rl.Rectangle{X: boxX, Y: g.Y + 18, Width: sz.X + 16, Height: 25}
	rl.DrawRectangleRounded(box, .25, 6, rl.Color{R: 77, G: 54, B: 170, A: 245})
	ui.DrawTextStyled(label, box.X+8, box.Y+6, 11, ui.FontSemiBold, rl.White)
}

func drawPassbandBracket(g rl.Rectangle, lowX, highX float32) {
	y := g.Y + g.Height - 13
	rl.DrawLineEx(rl.Vector2{X: lowX + 9, Y: y}, rl.Vector2{X: highX - 9, Y: y}, 2, colors.violet)
	rl.DrawTriangle(rl.Vector2{X: lowX, Y: y}, rl.Vector2{X: lowX + 10, Y: y - 5}, rl.Vector2{X: lowX + 10, Y: y + 5}, colors.violet)
	rl.DrawTriangle(rl.Vector2{X: highX, Y: y}, rl.Vector2{X: highX - 10, Y: y + 5}, rl.Vector2{X: highX - 10, Y: y - 5}, colors.violet)
}

func notchEdges(g rl.Rectangle, cfg audio.Config) (center, left, right float32) {
	center = spectrumX(g, cfg.Notch.FrequencyHz)
	halfPixels := float32(cfg.Notch.WidthHz/maxSpectrumHz) * g.Width * .5
	// A minimum visual width keeps both edge handles accessible even for a very
	// narrow real notch. Drag calculations still use the true frequency width.
	halfPixels = max(9, halfPixels)
	return center, max(g.X, center-halfPixels), min(g.X+g.Width, center+halfPixels)
}

func (a *application) drawNotchOverlay(g rl.Rectangle, cfg audio.Config) {
	center, left, right := notchEdges(g, cfg)
	amber := colors.amber
	bandColor := rl.Color{R: 244, G: 174, B: 28, A: 38}
	if !cfg.Notch.Enabled {
		amber = rl.Color{R: 127, G: 135, B: 157, A: 255}
		bandColor = rl.Color{R: 92, G: 99, B: 126, A: 25}
	}
	rl.DrawRectangleRec(rl.Rectangle{X: left, Y: g.Y, Width: right - left, Height: g.Height}, bandColor)
	if !cfg.Notch.Enabled {
		phase := float32(math.Mod(rl.GetTime()*22, 18))
		rl.BeginScissorMode(int32(left), int32(g.Y), int32(max(1, right-left)), int32(g.Height))
		for y := g.Y - g.Width + phase; y < g.Y+g.Height+g.Width; y += 18 {
			rl.DrawLineEx(rl.Vector2{X: left, Y: y}, rl.Vector2{X: right, Y: y + right - left}, 3, rl.Color{R: 15, G: 17, B: 24, A: 190})
		}
		rl.EndScissorMode()
	}
	rl.DrawLineEx(rl.Vector2{X: center, Y: g.Y}, rl.Vector2{X: center, Y: g.Y + g.Height}, 2.5, amber)
	drawGraphHandle(center, g.Y+10, amber, a.graphDrag == dragNotchCenter)
	handleY := g.Y + 205
	rl.DrawLineEx(rl.Vector2{X: left, Y: handleY}, rl.Vector2{X: left, Y: g.Y + g.Height}, 2, amber)
	rl.DrawLineEx(rl.Vector2{X: right, Y: handleY}, rl.Vector2{X: right, Y: g.Y + g.Height}, 2, amber)
	rl.DrawLineEx(rl.Vector2{X: left, Y: handleY}, rl.Vector2{X: right, Y: handleY}, 2, amber)
	drawGraphHandle(left, handleY, amber, a.graphDrag == dragNotchLeft)
	drawGraphHandle(right, handleY, amber, a.graphDrag == dragNotchRight)
	widthText := fmt.Sprintf("%.0f Hz", cfg.Notch.WidthHz)
	widthSize := ui.MeasureTextStyled(widthText, 11, ui.FontMono)
	mono(widthText, center-widthSize.X/2, handleY+14, 11, amber)
	label := "NOTCH · " + formatHz(cfg.Notch.FrequencyHz)
	sz := ui.MeasureTextStyled(label, 11, ui.FontSemiBold)
	boxX := max(g.X, min(g.X+g.Width-sz.X-16, center-sz.X/2-8))
	// Cutoff labels occupy the first row. Keep the notch on a dedicated second
	// row so nearby frequencies never make their boxes collide.
	box := rl.Rectangle{X: boxX, Y: g.Y + 51, Width: sz.X + 16, Height: 25}
	rl.DrawRectangleRounded(box, .25, 6, rl.Color{R: 126, G: 79, B: 10, A: 245})
	ui.DrawTextStyled(label, box.X+8, box.Y+6, 11, ui.FontSemiBold, rl.White)
	if a.graphDrag == dragNotchLeft || a.graphDrag == dragNotchRight {
		drawHint(center+14, handleY-11, "Arrastrar: ancho")
	} else if a.graphDrag == dragNotchCenter {
		drawHint(center+14, g.Y+83, "Arrastrar: frecuencia")
	}
}

func drawGraphHandle(x, y float32, accent rl.Color, active bool) {
	size := float32(14)
	if active {
		size = 17
	}
	r := rl.Rectangle{X: x - size*.5, Y: y - size*.5, Width: size, Height: size}
	// Dark fill separates the handle from both traces and tinted filter areas;
	// the white outer line remains visible at every spectrum level.
	rl.DrawRectangleRounded(r, .14, 4, colors.background)
	rl.DrawRectangleRoundedLinesEx(r, .14, 4, 3, rl.White)
	inner := rl.Rectangle{X: r.X + 3, Y: r.Y + 3, Width: r.Width - 6, Height: r.Height - 6}
	rl.DrawRectangleRec(inner, accent)
}

func drawHint(x, y float32, label string) {
	sz := ui.MeasureText(label, 11)
	box := rl.Rectangle{X: x, Y: y, Width: sz.X + 14, Height: 24}
	rl.DrawRectangleRounded(box, .2, 5, rl.Color{R: 11, G: 13, B: 25, A: 245})
	rl.DrawRectangleRoundedLinesEx(box, .2, 5, 1, colors.border)
	text(label, x+7, y+6, 11, colors.text)
}
func legend(x, y float32, c rl.Color, label string) {
	rl.DrawLineEx(rl.Vector2{X: x, Y: y + 7}, rl.Vector2{X: x + 28, Y: y + 7}, 3, c)
	text(label, x+36, y, 12, colors.text)
}
func drawTrace(g rl.Rectangle, data []float32, c rl.Color) {
	if len(data) < 2 {
		return
	}
	prev := rl.Vector2{X: g.X, Y: g.Y + g.Height*(1-max(0, min(1, (data[0]+80)/80)))}
	for i := 1; i < len(data); i++ {
		p := rl.Vector2{X: g.X + g.Width*float32(i)/float32(len(data)-1), Y: g.Y + g.Height*(1-max(0, min(1, (data[i]+80)/80)))}
		rl.DrawLineEx(prev, p, 1.5, c)
		prev = p
	}
}

func (a *application) handleSpectrumInput() {
	g := rl.Rectangle{X: 280, Y: 205, Width: 880, Height: 370}
	p := ui.MousePosition()
	over := ui.MouseInViewport() && rl.CheckCollisionPointRec(p, g) && !ui.PointerInputBlocked()
	if over && rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		cfg := a.configSnapshot()
		lowX, highX := spectrumX(g, cfg.LowCutHz), spectrumX(g, cfg.HighCutHz)
		center, left, right := notchEdges(g, cfg)
		switch {
		case cfg.Notch.Enabled && absf(p.X-left) <= 11 && p.Y >= g.Y+182 && p.Y <= g.Y+228:
			a.graphDrag = dragNotchLeft
		case cfg.Notch.Enabled && absf(p.X-right) <= 11 && p.Y >= g.Y+182 && p.Y <= g.Y+228:
			a.graphDrag = dragNotchRight
		case cfg.Notch.Enabled && absf(p.X-center) <= 10:
			if a.configSnapshot().Notch.Enabled {
				a.graphDrag = dragNotchCenter
			} else {
				a.graphDrag = dragNone
			}
		case absf(p.X-lowX) <= 10:
			a.graphDrag = dragLowCut
		case absf(p.X-highX) <= 10:
			a.graphDrag = dragHighCut
		default:
			now := rl.GetTime()
			if now-a.lastClick < .32 {
				a.update(func(c *audio.Config) { c.Notch.Enabled = !c.Notch.Enabled })
				a.notchOn.SetActive(a.configSnapshot().Notch.Enabled)
			}
			a.lastClick = now
			a.graphDrag = dragNotchCenter
		}
	}
	if a.graphDrag != dragNone && rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		hz := spectrumHz(g, p.X)
		cfg := a.configSnapshot()
		switch a.graphDrag {
		case dragLowCut:
			v := max(20, min(cfg.HighCutHz-50, hz))
			a.lowCut.SetValue(float32(v))
			a.update(func(c *audio.Config) { c.LowCutHz = v })
		case dragHighCut:
			v := max(cfg.LowCutHz+50, min(maxSpectrumHz, hz))
			a.highCut.SetValue(float32(v))
			a.update(func(c *audio.Config) { c.HighCutHz = v })
		case dragNotchCenter:
			v := max(80, min(maxSpectrumHz, hz))
			a.notchFreq.SetValue(float32(v))
			a.update(func(c *audio.Config) { c.Notch.FrequencyHz = v })
		case dragNotchLeft, dragNotchRight:
			v := max(20, min(2000, 2*math.Abs(hz-cfg.Notch.FrequencyHz)))
			a.notchWidth.SetValue(float32(v))
			a.update(func(c *audio.Config) { c.Notch.WidthHz = v })
		}
	}
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		a.graphDrag = dragNone
	}
	if over && a.configSnapshot().Notch.Enabled {
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				v := max(-60, min(-6, a.configSnapshot().Notch.DepthDB+float64(wheel)))
				a.notchDepth.SetValue(float32(v))
				a.update(func(c *audio.Config) { c.Notch.DepthDB = v })
			} else {
				v := max(20, min(2000, a.configSnapshot().Notch.WidthHz+float64(wheel)*10))
				a.notchWidth.SetValue(float32(v))
				a.update(func(c *audio.Config) { c.Notch.WidthHz = v })
			}
		}
	}
}

func (a *application) detectTone() {
	hz, ok := audio.DetectTone(a.before, maxSpectrumHz, 200, 6000)
	if !ok {
		a.message = "No se ha detectado un tono estrecho"
		return
	}
	a.notchFreq.SetValue(float32(hz))
	a.notchOn.SetActive(true)
	a.update(func(c *audio.Config) { c.Notch.Enabled = true; c.Notch.FrequencyHz = hz })
	a.message = fmt.Sprintf("Tono detectado: %.0f Hz", hz)
}
func (a *application) reset() {
	a.mu.Lock()
	a.config = audio.DefaultConfig()
	cfg := a.config
	a.mu.Unlock()
	a.processor.Configure(cfg)
	a.syncControls()
	if a.activeSlot >= 0 {
		a.presetDirty = cfg != a.activePresetConfig
	}
	a.message = "Ajustes restablecidos"
}
func (a *application) syncControls() {
	c := a.configSnapshot()
	a.lowCut.SetValue(float32(c.LowCutHz))
	a.highCut.SetValue(float32(c.HighCutHz))
	a.notchOn.SetActive(c.Notch.Enabled)
	a.notchFreq.SetValue(float32(c.Notch.FrequencyHz))
	a.notchWidth.SetValue(float32(c.Notch.WidthHz))
	a.notchDepth.SetValue(float32(c.Notch.DepthDB))
	a.noiseOn.SetActive(c.NoiseEnabled)
	a.blankerOn.SetActive(c.NoiseBlankerEnabled)
	a.noiseStrength.SetValue(float32(c.NoiseStrength))
	a.outputGain.SetValue(float32(c.OutputGainDB))
	a.processedMix.SetValue(float32(c.ProcessedMix))
	a.compressorOn.SetActive(c.Compressor.Enabled)
	a.compressorTarget.SetValue(float32(c.Compressor.TargetDBFS))
	a.ssbClarity.SetValue(float32(c.SSBClarityDB))
	a.eqOn.SetActive(c.EQEnabled)
	for i := range a.eq {
		a.eq[i].SetValue(float32(c.EQGainsDB[i]))
	}
	a.updateControlStates(c)
}
func presetPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "ic-dsp-preset.json"
	}
	return filepath.Join(dir, "IC-DSP", "preset.json")
}
func presetsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "ic-dsp-presets.json"
	}
	return filepath.Join(dir, "IC-DSP", "presets.json")
}
func settingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "ic-dsp-settings.json"
	}
	return filepath.Join(dir, "IC-DSP", "settings.json")
}
func (a *application) loadSettings() {
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return
	}
	settings := appSettings{ActivePreset: -1}
	if json.Unmarshal(data, &settings) == nil {
		a.preferredInput, a.preferredOutput = settings.InputDevice, settings.OutputDevice
		a.activeSlot = settings.ActivePreset
		a.themeMode = settings.Theme
	}
	if a.themeMode != "light" && a.themeMode != "dark" {
		a.themeMode = "auto"
	}
}
func (a *application) saveSettings() {
	path := settingsPath()
	if os.MkdirAll(filepath.Dir(path), 0755) != nil {
		return
	}
	data, err := json.MarshalIndent(appSettings{InputDevice: a.preferredInput, OutputDevice: a.preferredOutput, ActivePreset: a.activeSlot, Theme: a.themeMode}, "", "  ")
	if err == nil {
		_ = os.WriteFile(path, data, 0644)
	}
}
func (a *application) savePresetStore() error {
	path := presetsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(presetStore{Version: 2, Slots: a.presetSlots}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	return nil
}
func (a *application) loadPreset() {
	data, err := os.ReadFile(presetsPath())
	if err == nil {
		var store presetStore
		if json.Unmarshal(data, &store) == nil {
			a.presetSlots = store.Slots
			for i := range a.presetSlots {
				if store.Version < 2 && a.presetSlots[i].Saved {
					a.presetSlots[i].Config.ProcessedMix = 100
				}
				if a.presetSlots[i].Saved && a.presetSlots[i].Config.Compressor.TargetDBFS == 0 {
					a.presetSlots[i].Config.Compressor.TargetDBFS = -18
				}
				if a.presetSlots[i].Config.Emphasis == audio.PreEmphasis {
					a.presetSlots[i].Config.Emphasis = audio.DeEmphasis
				}
			}
			if a.activeSlot >= 0 && a.activeSlot < len(a.presetSlots) && a.presetSlots[a.activeSlot].Saved {
				a.config = a.presetSlots[a.activeSlot].Config
				a.activePresetConfig = a.config
			} else {
				a.activeSlot = -1
			}
			return
		}
	}
	// One-time migration from the original single-preset format.
	data, err = os.ReadFile(presetPath())
	if err != nil {
		return
	}
	var c audio.Config
	if json.Unmarshal(data, &c) != nil {
		return
	}
	if c.Compressor.TargetDBFS == 0 {
		c.Compressor.TargetDBFS = -18
	}
	c.ProcessedMix = 100
	if c.Emphasis == audio.PreEmphasis {
		c.Emphasis = audio.DeEmphasis
	}
	a.presetSlots[0] = presetSlot{Name: "Preset anterior", Saved: true, Config: c}
	a.activeSlot, a.config, a.activePresetConfig = 0, c, c
	_ = a.savePresetStore()
}

func (a *application) updatePresetButtons() {
	for i, button := range a.presetLoad {
		if button != nil {
			button.SetEnabled(a.presetSlots[i].Saved)
		}
	}
}

func (a *application) savePresetSlot(slot int) {
	if a.presetSlots[slot].Saved && (a.pendingOverwrite != slot || rl.GetTime() > a.overwriteUntil) {
		if a.pendingOverwrite >= 0 {
			a.presetSave[a.pendingOverwrite].SetLabel("GUARDAR")
		}
		a.pendingOverwrite = slot
		a.overwriteUntil = rl.GetTime() + 3
		a.presetSave[slot].SetLabel("CONFIRMAR")
		a.message = fmt.Sprintf("Confirma la sobrescritura del preset %d", slot+1)
		return
	}
	name := strings.TrimSpace(a.presetNames[slot].Text())
	if name == "" {
		name = fmt.Sprintf("Preset %d", slot+1)
		a.presetNames[slot].SetText(name)
	}
	cfg := a.configSnapshot()
	a.presetSlots[slot] = presetSlot{Name: name, Saved: true, Config: cfg}
	a.activeSlot, a.activePresetConfig, a.presetDirty = slot, cfg, false
	a.presetSave[slot].SetLabel("GUARDAR")
	a.pendingOverwrite = -1
	if err := a.savePresetStore(); err != nil {
		a.message = err.Error()
		return
	}
	a.saveSettings()
	a.updatePresetButtons()
	a.message = fmt.Sprintf("Preset %d guardado", slot+1)
}

func (a *application) loadPresetSlot(slot int) {
	if slot < 0 || slot >= len(a.presetSlots) || !a.presetSlots[slot].Saved {
		return
	}
	c := a.presetSlots[slot].Config
	if c.Compressor.TargetDBFS == 0 {
		c.Compressor.TargetDBFS = -18
	}
	if c.Emphasis == audio.PreEmphasis {
		c.Emphasis = audio.DeEmphasis
	}
	a.mu.Lock()
	a.config = c
	a.mu.Unlock()
	a.activeSlot, a.activePresetConfig, a.presetDirty = slot, c, false
	a.processor.Configure(c)
	a.syncControls()
	a.saveSettings()
	a.message = fmt.Sprintf("Preset %d cargado", slot+1)
}
