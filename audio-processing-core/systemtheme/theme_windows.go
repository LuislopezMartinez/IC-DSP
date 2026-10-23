//go:build windows

package systemtheme

import "golang.org/x/sys/windows/registry"

type Theme struct {
	Light            bool
	Red, Green, Blue uint8
}

func Read() Theme {
	result := Theme{Red: 124, Green: 92, Blue: 252}
	if key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE); err == nil {
		if value, _, err := key.GetIntegerValue("AppsUseLightTheme"); err == nil {
			result.Light = value != 0
		}
		_ = key.Close()
	}
	if key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\DWM`, registry.QUERY_VALUE); err == nil {
		if value, _, err := key.GetIntegerValue("ColorizationColor"); err == nil {
			result.Red = uint8(value)
			result.Green = uint8(value >> 8)
			result.Blue = uint8(value >> 16)
		}
		_ = key.Close()
	}
	return result
}
