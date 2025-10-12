# Testing Mode Guide

## Enable Testing Mode

Perfect for developers who are "shit at solving puzzles" (your words! 😄)

### Quick Enable

Edit `config.yaml`:

```yaml
puzzle:
  testing_mode: true
```

Then restart:
```bash
podman restart sliding-puzzle
```

## What It Does

When `testing_mode: true`:
- ✨ **Green glowing border** appears on the tile you SHOULD click next
- 👆 **Pointing finger emoji** shows exactly which tile to click
- 🔄 **Pulsing + bouncing animation** makes it impossible to miss
- 🤖 **Auto-Solve button** appears - click it to solve the puzzle automatically!
- 🎯 **Perfect solver** - hints always show the correct move from the recorded solution
- ✅ **Guaranteed solvable** - every puzzle is created by recording N moves from solved state

This lets you quickly complete puzzles without actually needing to solve them!

### Two Ways to Test:

1. **Manual with hints**: Click the tile with the 👆 finger (shows exact solution step-by-step)
2. **Auto-solve**: Click the "🤖 Auto-Solve" button (fastest for testing completion flow)

## Visual Effect

The correct tile to click will have:
- Bright green border (`#00ff00`)
- Glowing shadow effect
- Smooth pulsing animation
- 👆 Bouncing finger emoji overlay
- Updates after each move to show the next step

## Production Use

**IMPORTANT**: Set to `false` before giving to real users!

```yaml
puzzle:
  testing_mode: false  # Production setting
```

## Testing Workflow

1. **Enable testing mode**
   ```yaml
   testing_mode: true
   ```

2. **Test features:**
   - Click the "🤖 Auto-Solve" button for instant completion
   - Or manually follow the 👆 finger for step-by-step solving
   - Test email submission
   - Verify discount code generation
   - Check database storage
   - Test different grid sizes (3x3, 4x4, 5x5 - all instant with auto-solve!)

3. **Disable for production**
   ```yaml
   testing_mode: false
   ```

4. **Rebuild and deploy**
   ```bash
   podman build -t sliding-puzzle:latest .
   ./stop.sh
   ./start.sh
   ```

## Other Testing Tips

### Make Puzzle Easier
```yaml
puzzle:
  grid_size: 2           # 2x2 is super easy
  time_limit: 0          # No time pressure
  testing_mode: true     # Show which tiles to click
  scramble_moves: 10     # Only 10 moves to solve
  auto_solve_speed: 20   # Faster animation
```

### Quick Test Setup
```yaml
puzzle:
  grid_size: 3
  discount_percent: 99
  time_limit: 0
  testing_mode: true
  scramble_moves: 10     # Super quick testing
  auto_solve_speed: 20   # 0.2 seconds total
```

### View Test Results
```bash
# See all test completions
podman exec -it sliding-puzzle sqlite3 /app/data/puzzle.db \
  "SELECT email, discount_code, moves, time_seconds FROM completions ORDER BY created_at DESC LIMIT 10;"

# Count total completions
podman exec -it sliding-puzzle sqlite3 /app/data/puzzle.db \
  "SELECT COUNT(*) as total_completions FROM completions;"
```

## Pro Tips

### Quick Test Workflow
1. Start puzzle
2. Click "🤖 Auto-Solve" button
3. Watch it solve automatically with animation
4. Enter test email
5. Get discount code
6. Repeat!

### How Auto-Solve Works
1. **Scramble System**: Every puzzle is created by making N random moves from the solved state
2. **Solution Recording**: The scramble moves are recorded and reversed to create a perfect solution
3. **Auto-Solve**: Replays the solution in reverse with smooth animation
4. **Guaranteed Success**: Always solves in exactly `scramble_moves` steps (default: 50)
5. **Configurable Speed**: Adjust `auto_solve_speed` in config.yaml (default: 50ms per move)

### Speed Testing
- Auto-solve runs at configurable speed (default 50ms per move)
- 50 scramble moves = 2.5 seconds to solve
- 10 scramble moves = 0.5 seconds to solve (great for quick testing!)
- Total time: **Exactly `scramble_moves × auto_solve_speed` milliseconds**

## Enjoy Testing! 🎮

Now you can test everything without being good at puzzles - just click one button!

