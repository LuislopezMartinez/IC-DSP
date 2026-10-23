# SimpleUI

Biblioteca de interfaz retained-mode basada en [raylib-go](https://github.com/gen2brain/raylib-go). Esta copia está desacoplada de GO-Zero y forma parte del módulo `audio-processing-core`.

Incluye botones, interruptores, indicadores, etiquetas, sliders, range sliders, desplegables, campos de texto, temas, fuentes embebidas, canvas lógico escalable, viewport, gestión de foco y overlays.

## Importación

```go
import (
    ui "github.com/LuislopezMartinez/IC-DSP/audio-processing-core/simpleui"
    rl "github.com/gen2brain/raylib-go/raylib"
)
```

## Ejemplo mínimo

```go
package main

import (
    ui "github.com/LuislopezMartinez/IC-DSP/audio-processing-core/simpleui"
    rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
    ui.SetMode(1280, 720, ui.Fit)
    ui.SetTitle("Mi aplicación de audio")
    ui.SetTargetFPS(60)

    button := ui.NewButton("start", 40, 40, 180, 44, "INICIAR", 16)
    button.OnClick(func() { println("iniciar") })
    ui.Add(button)

    ui.SetColors(
        rl.Color{R: 10, G: 18, B: 24, A: 255},
        rl.Color{R: 3, G: 7, B: 10, A: 255},
    )
    ui.Run(func() {
        // Dibuje aquí gráficos propios. Run actualiza y dibuja los controles.
    })
}
```

## Traducción opcional

La biblioteca no depende del catálogo de idiomas de GO-Zero. El programa anfitrión puede instalar su propia función antes de `Run`:

```go
ui.SetTranslator(func(text string) string {
    if translated, ok := catalog[text]; ok {
        return translated
    }
    return text
}, []rune("áéíóúñü¿¡"))
```

Los codepoints adicionales se incorporan al cargar las fuentes embebidas.

## Recursos y licencias

Las fuentes Inter y JetBrains Mono están embebidas en `assets/fonts`. Sus licencias OFL se distribuyen junto a los archivos TTF. No es necesario copiar recursos en tiempo de ejecución.

## Notas

- Las funciones globales de configuración deben llamarse antes de `Run`.
- Las coordenadas de los controles pertenecen al canvas lógico configurado con `SetMode`.
- `Fit` conserva la relación de aspecto; consulte `viewport.go` para los demás modos.
- Los controles pueden administrarse con el manager global o con una instancia creada mediante `NewManager`.
- Raylib crea una ventana nativa, por lo que las pruebas automatizadas incluidas se concentran en lógica que no necesita abrirla.
