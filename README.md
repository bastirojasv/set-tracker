# Bear Tracker

Aplicación de escritorio para Windows que permite registrar y seguir el historial de sets en partidas de tenis u otros deportes de raqueta.

---

## Español

### ¿Qué hace?

Bear Tracker te permite llevar un registro detallado de tus partidas:

- **Tres modos de set:** Primero en llegar a N juegos (FT), Mejor de N (BO) o Libre (LS)
- **Historial completo** de todos los sets jugados
- **Estadísticas cara a cara (H2H)** filtradas por adversario
- **Perfil de usuario** con nombre, color y foto opcional
- Los datos se guardan localmente en una base de datos SQLite (`%APPDATA%\SetTracker\settracker.db`)

### Requisitos

- Windows 10/11 (64-bit)
- [Go 1.22+](https://go.dev/dl/)
- [MSYS2 MinGW-w64](https://www.msys2.org/) (para CGO y windres)

### Compilar

**Build completo con ícono embebido (recomendado):**
```powershell
./scripts/do_build.ps1
```

**Build simple:**
```
scripts/build.bat
```

**Build manual:**
```bash
go build -ldflags="-H windowsgui -s -w" -o BearTracker.exe .
```

### Estructura del proyecto

```
├── assets/          # Imágenes e ícono (bear.png, bear.ico, icon.rc)
├── scripts/         # Scripts de compilación (build.bat, do_build.ps1)
├── main.go          # Punto de entrada
├── app.go           # Lógica de navegación y pantallas
├── db.go            # Capa de acceso a SQLite
├── models.go        # Estructuras de datos
├── screen_*.go      # Una pantalla por archivo
├── widgets.go       # Widgets personalizados de Fyne
├── theme.go         # Tema y colores de la app
└── rsrc.syso        # Recurso Windows compilado (ícono en el .exe)
```

---

## English

### What does it do?

Bear Tracker lets you keep a detailed record of your matches:

- **Three set modes:** First to N games (FT), Best of N (BO), or Free/Open (LS)
- **Full history** of all played sets
- **Head-to-head (H2H) stats** filtered by opponent
- **User profile** with name, color, and optional photo
- Data is stored locally in a SQLite database (`%APPDATA%\SetTracker\settracker.db`)

### Requirements

- Windows 10/11 (64-bit)
- [Go 1.22+](https://go.dev/dl/)
- [MSYS2 MinGW-w64](https://www.msys2.org/) (for CGO and windres)

### Build

**Full build with embedded icon (recommended):**
```powershell
./scripts/do_build.ps1
```

**Simple build:**
```
scripts/build.bat
```

**Manual build:**
```bash
go build -ldflags="-H windowsgui -s -w" -o BearTracker.exe .
```

### Project structure

```
├── assets/          # Images and icon (bear.png, bear.ico, icon.rc)
├── scripts/         # Build scripts (build.bat, do_build.ps1)
├── main.go          # Entry point
├── app.go           # Navigation and screen logic
├── db.go            # SQLite data access layer
├── models.go        # Data structs
├── screen_*.go      # One file per screen
├── widgets.go       # Custom Fyne widgets
├── theme.go         # App theme and colors
└── rsrc.syso        # Compiled Windows resource (icon in the .exe)
```

---

Built with [Go](https://go.dev/) · [Fyne](https://fyne.io/) · [SQLite](https://pkg.go.dev/modernc.org/sqlite)
