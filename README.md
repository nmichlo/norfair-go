# 🎯 norfair-go

**Real-time multi-object tracking for Go**

[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://go.dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/nmichlo/norfair-go)](https://goreportcard.com/report/github.com/nmichlo/norfair-go)

---

> **⚠️ Disclaimer:** This is an unofficial Go port of Python's [norfair](https://github.com/tryolabs/norfair) object tracking library. This project is **NOT** affiliated with, endorsed by, or associated with Tryolabs or the original norfair development team. All credit for the original design and algorithms goes to the original norfair authors.

---

## Overview

**norfair-go** is a Go implementation of the norfair multi-object tracking library, bringing real-time object tracking capabilities to Go applications with:

- **Detector-agnostic design:** Works with any object detector (YOLO, Faster R-CNN, custom models)
- **Go-native performance:** Compiled binary with no Python overhead
- **Type-safe API:** Compile-time validation of tracking configurations
- **Comprehensive Tests:** Comprehensive test coverage and benchmarks ensuring 1:1 equivalence with the original norfair library

## Features

- 🔍 **Flexible Distance Functions:** IoU, Euclidean, Manhattan, Cosine, Custom Functions, and more+
- 🎛️ **Multiple Filtering Options:** Optimized Kalman filter, full filterpy-equivalent Kalman, or no filtering
- 🎥 **Video I/O:** Read/write video files with progress tracking and OpenCV integration
- 🎨 **Visualization:** Draw bounding boxes, keypoints, and motion trails with customizable colors
- 📹 **Camera Motion Compensation:** Support for translation, homography, and optical flow-based transformations
- 🔄 **Re-identification:** Optional feature embedding for robust identity matching

## Installation

```bash
go get github.com/nmichlo/norfair-go
```

**Note:** This library depends on [gocv](https://gocv.io) (Go bindings for OpenCV). You'll need to install OpenCV on your system:

```bash
# macOS
brew install opencv
# Ubuntu/Debian
sudo apt-get install libopencv-dev
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/nmichlo/norfair-go/pkg/norfairgo"
    "gonum.org/v1/gonum/mat"
)

func main() {
    // 1. Create tracker with IoU distance function
    tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
        DistanceFunction:    norfairgo.DistanceByName("iou"),
        DistanceThreshold:   0.5,
        HitCounterMax:       30,
        InitializationDelay: 3,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 2. for each timestep/frame
    for frame := range iterVideoFrames() {
        // 2.1 Generate detections from your object detector (e.g., D-FINE) or other source
        var detections []*norfairgo.Detection
        for _, bbox := range detectObjects(frame) {
            det, _ := norfairgo.NewDetection(
                mat.NewDense(2, 2, []float64{
                    bbox.X, bbox.Y,           // top-left
                    bbox.X+bbox.W, bbox.Y+bbox.H, // bottom-right
                }),
                nil, // optional: scores, labels, embeddings
            )
            detections = append(detections, det)
        }

        // 2.2 Update tracker, returning current tracked objects with stable IDs
        trackedObjects := tracker.Update(detections, 1, nil)

        // 2.3 Use tracked objects (draw, analyze, etc.)
        for _, obj := range trackedObjects {
            drawBox(frame, obj.Estimate, obj.ID)
        }
    }
}
```

<details>
<summary><b>🐍 Python Norfair Equivalent</b></summary>

Here's how the same tracking workflow looks in the original Python norfair library:

**Python:**
```python
from norfair import Detection, Tracker, Video, draw_tracked_objects

# Create tracker
tracker = Tracker(
    distance_function="iou",
    distance_threshold=0.5,
    hit_counter_max=30,
    initialization_delay=3,
)

# Process video
video = Video(input_path="input.mp4", output_path="output.mp4")

for frame in video:
    # Get detections from your detector
    yolo_detections = yolo_model(frame)

    # Convert to norfair detections
    detections = [
        Detection(points=bbox, scores=conf, label=cls)
        for bbox, conf, cls in yolo_detections
    ]

    # Update tracker
    tracked_objects = tracker.update(detections=detections)

    # Draw tracked objects
    draw_tracked_objects(frame, tracked_objects)
    video.write(frame)
```

**Key Differences:**
- **Go:** Explicit configuration structs vs Python kwargs
- **Go:** Error handling with `(result, error)` returns
- **Go:** Uses `gonum/mat` matrices instead of numpy arrays
- **Go:** Separate drawing package with explicit options
- **Go:** Type-safe API with compile-time validation

Both implementations provide the same core functionality with similar performance characteristics.

</details>

## Examples

This repository includes several working examples in the [`examples/`](examples/) directory:

- **`simple/`** - Basic tracking with simulated detections
- _More examples coming soon..._

Since functionality is intended to mirror the original norfair library, you can also refer to the original Python examples for guidance:

- [original norfair examples](https://github.com/tryolabs/norfair/tree/master/demos).

For a more advanced Go example demonstrating video processing, YOLO detection integration, and motion path visualization, see below:

<details>
<summary><b>🛠 Detailed Go Example</b></summary>

```go
package main

import (
    "fmt"
    "log"
    "github.com/nmichlo/norfair-go/pkg/drawing"
    "github.com/nmichlo/norfair-go/pkg/norfairgo"
    "gocv.io/x/gocv"
    "gonum.org/v1/gonum/mat"
)

func main() {
    // Configure video input
    video, err := norfairgo.NewVideo(&norfairgo.VideoOptions{
        InputPath:  "input.mp4",
        OutputPath: "output.mp4",
    })
    if err != nil {
        log.Fatalf("Failed to open video: %v", err)
    }
    defer video.Release()

    // Create tracker with advanced configuration
    tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
        DistanceFunction:       norfairgo.DistanceByName("iou"),
        DistanceThreshold:      0.5,
        HitCounterMax:          30,  // Keep tracking for 30 frames without detection
        InitializationDelay:    3,   // Require 3 detections to initialize
        PointwiseHitCounterMax: 4,   // Per-point tracking threshold
        DetectionThreshold:     0.5, // Minimum detection confidence
        PastDetectionsLength:   4,   // Store last 4 detections for reid
        FilterFactory:          norfairgo.OptimizedKalmanFilterFactory{},
    })
    if err != nil {
        log.Fatalf("Failed to create tracker: %v", err)
    }

    // Initialize motion path drawer
    paths := drawing.NewPaths(30) // Keep 30 frame history

    fmt.Println("Processing video...")
    frameNum := 0

    // Process each frame
    for {
        frame, err := video.Read()
        if err != nil {
            break // End of video
        }

        // Run your object detector (YOLO, etc.)
        detectionResults := runYOLODetector(frame)

        // Convert detector output to norfair detections
        var detections []*norfairgo.Detection
        for _, result := range detectionResults {
            det, err := norfairgo.NewDetection(
                mat.NewDense(2, 2, []float64{
                    result.BBox.X, result.BBox.Y,
                    result.BBox.X+result.BBox.Width, result.BBox.Y+result.BBox.Height,
                }),
                &norfairgo.DetectionConfig{
                    Scores: []float64{result.Confidence},
                    Label:  result.Class,
                },
            )
            if err != nil {
                log.Printf("Warning: Failed to create detection: %v", err)
                continue
            }
            detections = append(detections, det)
        }

        // Update tracker
        trackedObjects := tracker.Update(detections, 1, nil)

        // Visualize tracked objects
        drawing.DrawBoxes(frame, trackedObjects, &drawing.DrawOptions{
            ColorStrategy: "by_id",    // Color by object ID
            Thickness:     2,
            DrawLabels:    true,
            DrawIDs:       true,
        })

        // Draw motion paths
        paths.Update(trackedObjects)
        drawing.DrawPaths(frame, paths, &drawing.PathDrawOptions{
            ColorStrategy: "by_id",
            Thickness:     2,
            Fading:        true, // Fade older path points
        })

        // Write output frame
        if err := video.Write(frame); err != nil {
            log.Printf("Warning: Failed to write frame: %v", err)
        }

        frameNum++
        if frameNum%100 == 0 {
            fmt.Printf("Processed %d frames...\n", frameNum)
        }
    }

    fmt.Println("✓ Processing complete!")
    fmt.Printf("Tracked objects across %d frames\n", frameNum)
}

// Placeholder for your object detector
func runYOLODetector(frame gocv.Mat) []DetectionResult {
    // Replace with your actual detector implementation
    return []DetectionResult{}
}

type DetectionResult struct {
    BBox       struct{ X, Y, Width, Height float64 }
    Confidence float64
    Class      string
}
```

</details>

## Documentation

Full API documentation is available at [pkg.go.dev/github.com/nmichlo/norfair-go](https://pkg.go.dev/github.com/nmichlo/norfair-go).

### Core Types

- **`Tracker`** - Main tracking engine that maintains object identities across frames
- **`Detection`** - Input from object detector (bounding boxes, keypoints, or arbitrary points)
- **`TrackedObject`** - Output object with stable ID, position estimate, and tracking metadata
- **`Video`** - Video I/O with progress tracking and codec selection
- **`drawing.*`** - Visualization utilities for rendering tracked objects

## License & Attribution

**norfair-go** is licensed under the [BSD 3-Clause License](LICENSE).

This Go port is based on the original [norfair](https://github.com/tryolabs/norfair) by [Tryolabs](https://tryolabs.com/) (BSD 3-Clause). Their well-designed, detector-agnostic architecture made this port possible. Internal packages include code adapted from several Python libraries—see [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md) for complete attribution.

**Citation:** If using this library in research, please cite the original norfair paper as described [here](https://github.com/tryolabs/norfair).

---

**Contributing:** Issues and pull requests welcome at [github.com/nmichlo/norfair-go](https://github.com/nmichlo/norfair-go)
