# TUI — Terminal User Interface

Interfaz interactiva construida con [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-Architecture) y [lipgloss](https://github.com/charmbracelet/lipgloss) para el renderizado con estilos.

## Arquitectura

El paquete sigue el patrón **Elm-Architecture** de Bubble Tea con tres métodos principales:

| Método   | Responsabilidad |
|----------|-----------------|
| `Init()` | Arranca el spinner y lanza el análisis del primer plugin. |
| `Update(msg)` | Procesa mensajes (resize, teclas, resultados de análisis). |
| `View()` | Renderiza el frame actual según el estado del modelo. |


## Pantallas

### 1. Loading

Se muestra mientras los plugins se analizan de forma secuencial. Contiene:

- **Spinner** animado con progreso (`3/25`).
- Nombre del plugin en análisis.
- Lista creciente de plugins ya completados con su icono de severidad.

### 2. Lista (`viewList`)

Pantalla principal con tabla scrollable:

```
  ⚡ NvimGoTrack — Plugin Breaking-Change Tracker

  ● 2 breaking  ● 1 deprecated  ● 5 behind  │  25 plugins total
  [ All ] [ 🔴 Breaking ] [ 🟡 Deprecated ] [ 📦 Behind ]

       Plugin                           Commit       Behind     Status
  ─────────────────────────────────────────────────────────────────────
  🔴 telescope.nvim                     a1b2c3d4e5   +12        BREAKING
  🟡 nvim-treesitter                    f6g7h8i9j0   +3         deprecated
  ✅ plenary.nvim                       k1l2m3n4o5              up to date
```

**Componentes:**
- **Barra de resumen** — conteo de breaking / deprecated / behind.
- **Pestañas de filtro** — `All`, `🔴 Breaking`, `🟡 Deprecated`, `📦 Behind`.
- **Tabla** — icono, nombre (max 30 chars), commit (max 10 chars), behind count, estado.
- **Auto-scroll** — la lista sigue al cursor dentro de la altura del terminal.

### 3. Detalle (`viewDetail`)

Información completa del plugin seleccionado:

- **Metadatos:** repositorio, branch, commit actual, commits detrás, severidad, URL de comparación.
- **🔴 Breaking Changes** — mensajes de commits con cambios incompatibles.
- **🟡 Deprecation Warnings** — mensajes de commits con deprecaciones.
- **📦 Recent Releases** — hasta 10 releases con tag, nombre, y snippet del body (3 líneas).

## Atajos de teclado

| Tecla | Vista Lista | Vista Detalle |
|-------|-------------|---------------|
| `j` / `↓` | Mover cursor abajo | Scroll abajo |
| `k` / `↑` | Mover cursor arriba | Scroll arriba |
| `g` | Ir al primer elemento | Scroll al inicio |
| `G` | Ir al último elemento | — |
| `Enter` | Abrir detalle | — |
| `Tab` | Siguiente filtro | Siguiente filtro |
| `Shift+Tab` | Filtro anterior | Filtro anterior |
| `Esc` | — | Volver a lista |
| `q` / `Ctrl+C` | Salir | Volver a lista |

## Filtros

El filtrado se controla con `Tab` / `Shift+Tab` y cicla entre 4 modos:

| Filtro | Descripción |
|--------|-------------|
| `filterAll` | Todos los plugins |
| `filterBreaking` | Solo `SeverityBreaking` |
| `filterDeprecated` | `SeverityDeprecation` o superior |
| `filterBehind` | Plugins con `BehindBy > 0` |

Al cambiar de filtro, el cursor se reinicia a `0` y la lista `filtered` se reconstruye.

## Mensajes internos (Bubble Tea)

| Mensaje | Origen | Efecto |
|---------|--------|--------|
| `tea.WindowSizeMsg` | Terminal | Actualiza `width` y `height` |
| `spinner.TickMsg` | Spinner | Anima el spinner durante la carga |
| `pluginAnalyzed` | `analyzeNext()` | Guarda el `PluginReport`, aplica filtro, lanza siguiente análisis |
| `allDone` | `analyzeNext()` | Detiene spinner, ordena reports por severidad |
| `tea.KeyMsg` | Teclado | Delega a `handleKey()` |

## Pipeline de análisis

Los plugins se analizan **secuencialmente** (uno a la vez) para respetar los rate limits de la API de GitHub:

```
Init() → analyzeNext(0) → pluginAnalyzed{0} → analyzeNext(1) → ... → allDone
```

Cada `pluginAnalyzed` actualiza `reports[i]`, incrementa `loadingIdx`, y reaplica el filtro para que la pantalla de loading se actualice en tiempo real.

## Estilos (`styles.go`)

Paleta de colores oscura con semántica de severidad:

| Variable | Hex | Uso |
|----------|-----|-----|
| `colorBreaking` | `#FF4444` | Breaking changes, errores |
| `colorDeprecation` | `#FFB020` | Deprecaciones |
| `colorFeature` | `#44DD88` | Plugins con actualizaciones |
| `colorOK` | `#88AACC` | Plugins al día |
| `colorAccent` | `#7C6FFF` | Títulos, tabs activos, tags de release |
| `colorBgSelected` | `#2A2B4E` | Fila seleccionada |
| `colorMuted` | `#666677` | Texto de ayuda, bordes |

## Estructura de archivos

```
internal/tui/
├── tui.go       # Model, Init, Update, View y toda la lógica de la TUI
└── styles.go    # Paleta de colores y estilos lipgloss
```

## Dependencias

| Paquete | Uso |
|---------|-----|
| `bubbles/spinner` | Animación de spinner durante la carga |
| `bubbletea` | Framework Elm-Architecture para TUIs |
| `lipgloss` | Estilos y colores del terminal |

