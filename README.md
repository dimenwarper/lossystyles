# lossystyles

Rich terminal visualizations for ML training metrics. Connect it like wandb, but everything stays in your terminal.

## Architecture

**Two components:**

- **Go CLI** (`cli/`) — Bubble Tea TUI dashboard with multiple themes
- **Python SDK** (`sdk/`) — `pip install lossystyles`, thin client that sends metrics over Unix socket

## Themes

Each theme is a genuinely different visual experience, not just a color swap:

| Theme | Style |
|-------|-------|
| `dark` | Clean dark terminal with braille charts |
| `neon` | Bright cyberpunk palette |
| `retro` | Warm amber CRT aesthetic |
| `minimal` | Stripped-down monochrome |
| `rainbow` | Cycling rainbow colors on everything |
| `bio` | Gel electrophoresis lanes with smear-to-bands separation, DNA helix sparklines, per-lane colors (white/green/blue/orange) |
| `eva` | NERV HUD with rotating icosahedron wireframe, sine wave background, block "02", MAGI system footer |

## Quick start

### Demo mode (no Python needed)

```bash
cd cli
go build ./cmd/lossystyles/
./lossystyles --demo --theme eva
```

### With the Python SDK

```bash
pip install -e sdk/
```

```python
import lossystyles

run = lossystyles.init(project="my-experiment", config={"lr": 3e-4})

for step in range(100):
    run.log({"loss": loss, "accuracy": acc}, step=step)

run.finish()
```

Then in another terminal:

```bash
./lossystyles --run-id <run-id> --theme bio
```

## Protocol

NDJSON over Unix socket (`/tmp/lossystyles-{run-id}.sock`). See [`protocol/PROTOCOL.md`](protocol/PROTOCOL.md) for the spec.
