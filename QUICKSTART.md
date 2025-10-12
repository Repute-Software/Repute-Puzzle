# Quick Start Guide

Get your puzzle marketing tool up and running in 3 minutes!

## Step 1: Add an Image

Copy at least one PNG image to the `images/` directory:

```bash
cp /path/to/your-image.png images/
```

## Step 2: Build the Container

```bash
podman build -t sliding-puzzle:latest .
```

## Step 3: Start the Container

```bash
./start.sh
```

## Step 4: Open the App

Visit http://localhost:8080 in your browser.

---

## Customization

Edit `config.yaml` to change puzzle settings:

```yaml
puzzle:
  grid_size: 4           # Change to 4x4 grid
  discount_percent: 20   # Change to 20% discount
  time_limit: 180        # Change to 3 minutes
  testing_mode: true     # Highlight clickable tiles (great for testing!)
```

**Pro tip:** Set `testing_mode: true` to see exactly which tile to click next (green glow + 👆 finger) - perfect for testing!

Then restart:

```bash
podman restart sliding-puzzle
```

---

## View Discount Codes

```bash
podman exec -it sliding-puzzle sqlite3 /app/data/puzzle.db \
  "SELECT email, discount_code, created_at FROM completions;"
```

---

## Stop the App

```bash
./stop.sh
```

---

For full documentation, see [README.md](README.md)

