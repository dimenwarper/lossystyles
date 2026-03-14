"""PyTorch / PyTorch Lightning callback for lossystyles."""

from __future__ import annotations

from typing import Any

import lossystyles


class LossyStylesCallback:
    """Callback that logs training metrics to lossystyles.

    Works with plain PyTorch training loops or PyTorch Lightning.

    Usage with plain PyTorch:
        cb = LossyStylesCallback(project="my-model")
        cb.on_train_begin()
        for step, batch in enumerate(loader):
            loss = train_step(batch)
            cb.on_step_end(step, {"loss": loss.item()})
        cb.on_train_end()

    Usage with PyTorch Lightning:
        trainer = Trainer(callbacks=[LossyStylesCallback()])
    """

    def __init__(
        self,
        project: str = "",
        theme: str = "dark",
        config: dict | None = None,
    ):
        self.project = project
        self.theme = theme
        self.config = config or {}

    def on_train_begin(self, **kwargs: Any) -> None:
        lossystyles.init(
            project=self.project,
            config=self.config,
            theme=self.theme,
        )

    def on_step_end(self, step: int, metrics: dict[str, float], **kwargs: Any) -> None:
        lossystyles.log(metrics, step=step)

    def on_train_end(self, **kwargs: Any) -> None:
        lossystyles.finish()
