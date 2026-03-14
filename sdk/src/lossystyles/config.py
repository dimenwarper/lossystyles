from dataclasses import dataclass, field


@dataclass
class RunConfig:
    project: str = ""
    config: dict = field(default_factory=dict)
    theme: str = "dark"
