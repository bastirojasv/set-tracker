@echo off
REM SetTracker — build script
REM Requiere: Go 1.22+, GCC via MSYS2 MinGW-w64

REM Agrega Go y GCC de MSYS2 al PATH si no están ya
set "GO_BIN=C:\Program Files\Go\bin"
set "MINGW_BIN=C:\msys64\mingw64\bin"
echo %PATH% | find /I "%GO_BIN%" >nul 2>&1
if errorlevel 1 set "PATH=%GO_BIN%;%PATH%"
echo %PATH% | find /I "%MINGW_BIN%" >nul 2>&1
if errorlevel 1 set "PATH=%MINGW_BIN%;%PATH%"

set GOARCH=amd64
set GOOS=windows
set CGO_ENABLED=1

go build -ldflags="-H windowsgui -s -w" -o SetTracker.exe .
if %ERRORLEVEL% == 0 (
    echo Build exitoso: SetTracker.exe
) else (
    echo Error en build.
)
