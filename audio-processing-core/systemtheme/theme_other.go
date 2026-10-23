//go:build !windows

package systemtheme

type Theme struct {
	Light            bool
	Red, Green, Blue uint8
}

func Read() Theme { return Theme{Red: 124, Green: 92, Blue: 252} }
