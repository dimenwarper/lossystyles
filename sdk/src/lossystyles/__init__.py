"""lossystyles — Rich CLI visualizations for ML training."""

from lossystyles.config import RunConfig
from lossystyles.run import Run

__all__ = ["init", "log", "finish", "Run", "RunConfig"]

_active_run: Run | None = None


def init(
    project: str = "",
    config: dict | None = None,
    theme: str = "dark",
    auto_launch: bool = True,
) -> Run:
    """Initialize a new lossystyles run.

    Args:
        project: Project name for display.
        config: Hyperparameters dict shown in dashboard.
        theme: Theme name (dark, neon, retro, minimal).
        auto_launch: Whether to auto-launch the CLI binary.

    Returns:
        A Run instance. Also usable as a context manager.
    """
    global _active_run
    run_config = RunConfig(project=project, config=config or {}, theme=theme)
    run = Run(run_config, auto_launch=auto_launch)
    run.start()
    _active_run = run
    return run


def log(metrics: dict[str, float], step: int | None = None) -> None:
    """Log metrics to the active run."""
    if _active_run is None:
        raise RuntimeError("No active run. Call lossystyles.init() first.")
    _active_run.log(metrics, step=step)


def finish() -> None:
    """Finish the active run and tear down the dashboard."""
    global _active_run
    if _active_run is None:
        return
    _active_run.finish()
    _active_run = None
