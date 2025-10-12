#!/bin/bash
# Start the sliding puzzle container

# Check if container is already running
if podman ps -a --format "{{.Names}}" | grep -q "^sliding-puzzle$"; then
    echo "Container 'sliding-puzzle' already exists. Removing it..."
    podman rm -f sliding-puzzle
fi

# Run the container
echo "Starting sliding-puzzle container..."
podman run -d \
  --name sliding-puzzle \
  -p 8080:8080 \
  -v "$(pwd)/images:/app/images:ro" \
  -v "$(pwd)/data:/app/data" \
  -v "$(pwd)/config.yaml:/app/config.yaml:ro" \
  --restart unless-stopped \
  localhost/sliding-puzzle:latest

echo ""
echo "✅ Container started successfully!"
echo ""
echo "🌐 Access the app at: http://localhost:8080"
echo ""
echo "📝 Useful commands:"
echo "  - View logs:  podman logs -f sliding-puzzle"
echo "  - Stop:       podman stop sliding-puzzle"
echo "  - Restart:    podman restart sliding-puzzle"
echo "  - Remove:     podman rm -f sliding-puzzle"
echo ""

