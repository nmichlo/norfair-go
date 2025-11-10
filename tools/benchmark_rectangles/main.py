#!/usr/bin/env python3
# /// script
# dependencies = ["norfair>=2.3.0", "numpy>=1.20.0"]
# ///
"""
Python benchmark using norfair for rectangle tracking performance comparison

Run this file with $ uv run tools/benchmark_rectangles/main.py
"""

import time
from typing import List, Tuple
import numpy as np
import norfair

# ========================================================================= #
# RNG
# ========================================================================= #


class SimpleRNG:
    """Simple RNG using LCG algorithm"""

    def __init__(self, seed: int):
        self.state = seed & 0xFFFFFFFF  # Keep it 32-bit

    def next(self) -> int:
        """Returns the next random uint32"""
        # LCG parameters from Numerical Recipes
        a = 1664525
        c = 1013904223
        m = 1 << 32

        self.state = (a * self.state + c) % m
        return self.state

    def random(self) -> float:
        """Returns a random float in [0.0, 1.0)"""
        return self.next() / (1 << 32)


# ========================================================================= #
# SIM
# ========================================================================= #



class Rectangle:
    """Represents a moving bounding box in the simulation"""

    def __init__(self, id: int, x: float, y: float, width: float, height: float, vx: float, vy: float):
        self.id = id  # Ground truth ID
        self.x = x  # Center X position
        self.y = y  # Center Y position
        self.width = width  # Box width
        self.height = height  # Box height
        self.vx = vx  # Velocity X (pixels per frame)
        self.vy = vy  # Velocity Y (pixels per frame)


class Simulation:
    """Manages the rectangle physics environment"""

    def __init__(self, width: int, height: int, num_rectangles: int, seed: int):
        self.width = width
        self.height = height
        self.rectangles: List[Rectangle] = []

        # Use deterministic RNG
        rng = SimpleRNG(seed)

        # Initialize rectangles with random positions and velocities
        for i in range(num_rectangles):
            rect = Rectangle(
                id=i + 1,  # IDs start at 1
                x=rng.random() * width,
                y=rng.random() * height,
                width=20.0 + rng.random() * 60.0,  # 20-80 pixels
                height=20.0 + rng.random() * 60.0,  # 20-80 pixels
                vx=-5.0 + rng.random() * 10.0,  # -5 to +5 pixels/frame
                vy=-5.0 + rng.random() * 10.0,  # -5 to +5 pixels/frame
            )
            self.rectangles.append(rect)

    def update(self):
        """Advance the simulation by one frame"""
        for rect in self.rectangles:
            # Update position
            rect.x += rect.vx
            rect.y += rect.vy

            # Bounce off walls (left/right)
            half_w = rect.width / 2.0
            if rect.x - half_w < 0:
                rect.x = half_w
                rect.vx = -rect.vx
            elif rect.x + half_w > self.width:
                rect.x = self.width - half_w
                rect.vx = -rect.vx

            # Bounce off walls (top/bottom)
            half_h = rect.height / 2.0
            if rect.y - half_h < 0:
                rect.y = half_h
                rect.vy = -rect.vy
            elif rect.y + half_h > self.height:
                rect.y = self.height - half_h
                rect.vy = -rect.vy

    def get_bounding_boxes(self) -> List[np.ndarray]:
        """Returns bounding boxes in [[x_min, y_min], [x_max, y_max]] format"""
        boxes = []

        for rect in self.rectangles:
            half_w = rect.width / 2.0
            half_h = rect.height / 2.0

            x_min = rect.x - half_w
            y_min = rect.y - half_h
            x_max = rect.x + half_w
            y_max = rect.y + half_h

            box = np.array([[x_min, y_min], [x_max, y_max]], dtype=np.float64)
            boxes.append(box)

        return boxes

    def get_ground_truth_ids(self) -> List[int]:
        """Returns the ground truth IDs for validation"""
        return [rect.id for rect in self.rectangles]


# ========================================================================= #
# MAIN
# ========================================================================= #

def run_benchmark(filter_name: str, filter_factory, num_objects: int, num_frames: int):
    """Run benchmark for given configuration"""

    # Create tracker
    tracker = norfair.Tracker(
        distance_function="iou",
        distance_threshold=0.5,
        hit_counter_max=30,
        initialization_delay=3,
        pointwise_hit_counter_max=4,
        detection_threshold=0.0,
        past_detections_length=4,
        filter_factory=filter_factory,
    )

    # Create simulation
    sim = Simulation(1920, 1080, num_objects, 42)

    # Warmup (10 frames)
    for _ in range(10):
        sim.update()
        boxes = sim.get_bounding_boxes()
        detections = [
            norfair.Detection(points=box)
            for box in boxes
        ]
        tracker.update(detections)

    # Benchmark
    total_time_sim = 0
    total_time_bench = 0
    for frame in range(num_frames):
        t0 = time.time_ns()
        sim.update()
        boxes = sim.get_bounding_boxes()

        # Create detections from bounding boxes
        t1 = time.time_ns()
        detections = [
            norfair.Detection(points=box)
            for box in boxes
        ]

        # Track objects
        tracked_objects = tracker.update(detections)

        t2 = time.time_ns()
        total_time_sim += (t1 - t0)
        total_time_bench += (t2 - t1)

    # Calculate metrics
    fps = num_frames / (total_time_bench / 1_000_000_000)
    avg_time_ms = (total_time_bench / num_frames) / 1_000_000

    print(f"Python - Filter: {filter_name:<15} Objects: {num_objects:3}  |  FPS: {fps:7.1f}  |  Avg: {avg_time_ms:6.3f}ms")


def main():
    num_frames = 1000

    # Test configurations
    filters = [
        ("OptimizedKalman", norfair.OptimizedKalmanFilterFactory()),
        ("FilterPyKalman", norfair.FilterPyKalmanFilterFactory()),
        ("NoFilter", norfair.NoFilterFactory()),
    ]

    object_counts = [10, 50, 100]

    for count in object_counts:
        print(f"\n--- {count} Objects, {num_frames} frames ---")
        for filter_name, filter_factory in filters:
            run_benchmark(filter_name, filter_factory, count, num_frames)


if __name__ == "__main__":
    main()
