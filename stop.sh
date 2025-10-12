#!/bin/bash
# Stop and remove the sliding puzzle container

echo "Stopping sliding-puzzle container..."
podman stop sliding-puzzle 2>/dev/null
podman rm sliding-puzzle 2>/dev/null

echo "✅ Container stopped and removed."

