"""Basic lossystyles example — simulates a training loop."""

import math
import random
import time

import lossystyles

run = lossystyles.init(
    project="gpt-micro",
    config={"lr": 3e-4, "batch_size": 64, "model": "transformer"},
    theme="dark",
)

total_steps = 300
base_loss = 2.5

for step in range(total_steps):
    progress = step / total_steps
    loss = base_loss * math.exp(-3 * progress) * (1 + 0.1 * random.random())
    accuracy = (1 - math.exp(-4 * progress)) * (0.95 + 0.05 * random.random())

    # LR warmup then cosine decay
    if step < 20:
        lr = 3e-4 * step / 20
    else:
        lr = 3e-4 * 0.5 * (1 + math.cos(math.pi * (step - 20) / (total_steps - 20)))

    lossystyles.log({"loss": loss, "accuracy": accuracy, "lr": lr}, step=step)
    time.sleep(0.05)

lossystyles.finish()
