# lossystyles IPC Protocol

Communication between the Python SDK and Go CLI uses **newline-delimited JSON (NDJSON)** over a **Unix domain socket**.

## Socket Path

Default: `/tmp/lossystyles-{run_id}.sock`

## Message Format

Each message is a single JSON object followed by `\n`:

```json
{"type": "init", "run_id": "abc123", "project": "gpt-micro", "config": {"lr": 3e-4}, "theme": "dark"}
{"type": "log", "run_id": "abc123", "step": 42, "metrics": {"loss": 0.234, "lr": 0.0003}}
{"type": "finish", "run_id": "abc123"}
```

## Message Types

### `init`
Starts a new run. The CLI opens the dashboard.

| Field     | Type   | Required | Description              |
|-----------|--------|----------|--------------------------|
| type      | string | yes      | `"init"`                 |
| run_id    | string | yes      | Unique run identifier    |
| project   | string | no       | Project name for display |
| config    | object | no       | Hyperparameters          |
| theme     | string | no       | Theme name (default: "dark") |

### `log`
Appends metrics for a step.

| Field   | Type   | Required | Description            |
|---------|--------|----------|------------------------|
| type    | string | yes      | `"log"`                |
| run_id  | string | yes      | Run identifier         |
| step    | int    | no       | Global step number     |
| metrics | object | yes      | Key-value metric pairs |

### `finish`
Ends the run and tears down the dashboard.

| Field  | Type   | Required | Description        |
|--------|--------|----------|--------------------|
| type   | string | yes      | `"finish"`         |
| run_id | string | yes      | Run identifier     |
