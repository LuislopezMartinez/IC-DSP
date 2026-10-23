$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$toolBin = Join-Path (Split-Path -Parent $projectRoot) 'GO-Zero\build\radiosonde-toolchain\w64devkit\bin'
$compiler = Join-Path $toolBin 'gcc.exe'
$cppCompiler = Join-Path $toolBin 'g++.exe'

if (-not (Test-Path -LiteralPath $compiler)) {
    throw "No se encontró el compilador de GO-Zero en: $compiler"
}

$env:PATH = "$toolBin;$env:PATH"
$env:CC = $compiler
$env:CXX = $cppCompiler
$env:CGO_ENABLED = '1'

Push-Location (Join-Path $projectRoot 'audio-processing-core')
try {
    go test ./...
    # Use the Windows GUI subsystem so launching the application does not open
    # an additional Command Prompt window behind the Raylib interface.
    go build -tags noaudio -trimpath -ldflags '-H=windowsgui' -o (Join-Path $projectRoot 'ic-dsp.exe') ./cmd/ic-dsp
}
finally {
    Pop-Location
}

Write-Host "Compilación completada: $(Join-Path $projectRoot 'ic-dsp.exe')"
