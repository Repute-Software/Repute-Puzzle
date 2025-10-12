# Sliding Puzzle Marketing Tool

A containerized web application that gamifies marketing by having users complete a sliding puzzle to earn discount codes. Built with Go, SQLite, Templ, and HTMX.

## Features

- 🧩 Configurable puzzle grid size (3x3, 4x4, 5x5, etc.)
- 🎁 Automatic discount code generation
- 📧 Email collection for marketing
- 🗄️ SQLite database for tracking completions
- 🐳 Fully containerized with Docker/Podman
- ⚡ Fast, responsive UI with HTMX
- 🎨 Modern, beautiful design
- 📱 Mobile-friendly responsive layout

## Prerequisites

- Podman (or Docker)
- Podman Compose (or Docker Compose)

## Quick Start

### 1. Clone or Setup the Project

Ensure you have all the files in your project directory.

### 2. Add Puzzle Images

Place one or more PNG images in the `images/` directory. These will be used for the puzzle:

```bash
cp your-image.png images/
```

**Recommended image specifications:**
- Format: PNG
- Size: 600x600 pixels or larger
- Square aspect ratio works best

### 3. Configure the Puzzle

Edit `config.yaml` to customize your puzzle:

```yaml
puzzle:
  grid_size: 3           # 3x3, 4x4, 5x5, etc. (2-10)
  discount_percent: 15   # Discount amount (0-100)
  time_limit: 300        # Time limit in seconds (0 = no limit)
  testing_mode: false    # Highlight clickable tiles for testing
  scramble_moves: 50     # Number of moves to scramble (1-1000)
  auto_solve_speed: 50   # Speed of auto-solve animation in ms (1-10000)
database:
  path: "./data/puzzle.db"
server:
  port: 8080
images:
  directory: "./images"
```

### 4. Build the Container

```bash
podman build -t sliding-puzzle:latest .
```

### 5. Run the Container

Use the provided helper script:

```bash
./start.sh
```

Or run manually:

```bash
podman run -d \
  --name sliding-puzzle \
  -p 8080:8080 \
  -v "$(pwd)/images:/app/images:ro" \
  -v "$(pwd)/data:/app/data" \
  -v "$(pwd)/config.yaml:/app/config.yaml:ro" \
  localhost/sliding-puzzle:latest
```

### 6. Access the Application

Open your browser and navigate to:

```
http://localhost:8080
```

## Production Deployment

For production deployment with pre-built images from GitHub Container Registry, see [DEPLOYMENT.md](DEPLOYMENT.md).

### Quick Production Start

Once the image is published to GitHub Container Registry, deploy on any server:

```bash
# Clone repository
git clone https://github.com/USERNAME/puzzle.git
cd puzzle

# Edit compose.prod.yaml with your GitHub username
vim compose.prod.yaml

# Configure settings and add images
vim config.yaml
cp your-images/*.png images/

# Deploy
docker compose -f compose.prod.yaml up -d
```

Images are automatically built and available at:
- `ghcr.io/USERNAME/puzzle:latest` - Latest build from main branch
- `ghcr.io/USERNAME/puzzle:v1.0.0` - Specific version tags

### Benefits of Production Deployment

✅ **No Build Required**: Pull pre-built images from GitHub  
✅ **Automatic Updates**: Images built on every push to main  
✅ **Version Control**: Use semantic versioning for stability  
✅ **Easy Rollback**: Switch to any previous version instantly  
✅ **CI/CD Pipeline**: Automated testing and deployment  

## Multi-Language Support

The application supports multiple languages via URL parameters:

- **English (default)**: `http://localhost:8080`
- **Dutch**: `http://localhost:8080?lang=nl`

Users can also switch languages using the flag selector (🇬🇧 EN | 🇳🇱 NL) at the top of the page.

To add more languages, add translation files in the `locales/` directory. See [I18N_COMPLETE.md](I18N_COMPLETE.md) for details.

## Configuration Options

### Puzzle Settings

- **grid_size**: The dimensions of the puzzle grid (e.g., 3 creates a 3x3 puzzle with 9 tiles)
  - Range: 2-10
  - Larger grids are more challenging

- **discount_percent**: The discount percentage shown in the code
  - Range: 0-100
  - Appears in the generated code format: `PUZZLE15-ABC123`

- **time_limit**: Time limit for completing the puzzle in seconds
  - Set to 0 for no time limit
  - Recommended: 180-600 seconds depending on grid size

- **testing_mode**: Adds testing features to help you complete puzzles easily
  - Set to `true` for testing - shows perfect hints and auto-solve button
  - The correct tile glows green with a 👆 pointing finger (shows exact solution)
  - Adds a "🤖 Auto-Solve" button that replays the solution with animation
  - Set to `false` for production use (hides all hints and auto-solve)
  - Perfect for testing the full flow without puzzle-solving skills!

- **scramble_moves**: Number of random moves to scramble the puzzle
  - Range: 1-1000 moves
  - Default: 50 moves
  - More moves = harder puzzle (takes longer for users to solve)
  - Fewer moves = easier puzzle (great for testing: try 10 moves)
  - Every puzzle is guaranteed solvable

- **auto_solve_speed**: Animation speed for auto-solve in milliseconds
  - Range: 1-10000 ms per move
  - Default: 50ms (smooth animation)
  - Lower = faster (20ms = very fast)
  - Higher = slower (100ms = easier to watch)
  - Total solve time = `scramble_moves × auto_solve_speed`

### Database

- **path**: Location of the SQLite database file
  - Default: `./data/puzzle.db`
  - Data persists across container restarts

### Server

- **port**: HTTP server port
  - Default: 8080
  - Change the port in both `config.yaml` and `compose.yaml`

## Database Structure

The application stores completion records in SQLite with the following schema:

```sql
CREATE TABLE completions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    discount_code TEXT NOT NULL UNIQUE,
    grid_size INTEGER NOT NULL,
    moves INTEGER NOT NULL,
    time_seconds INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Viewing Completion Data

To view all completions and discount codes issued:

```bash
# Access the database
podman exec -it sliding-puzzle sqlite3 /app/data/puzzle.db

# Run queries
sqlite> SELECT email, discount_code, moves, time_seconds, created_at 
        FROM completions 
        ORDER BY created_at DESC;

sqlite> .exit
```

Or export to CSV:

```bash
podman exec -it sliding-puzzle sqlite3 -csv /app/data/puzzle.db \
  "SELECT * FROM completions;" > completions.csv
```

## Managing the Container

### View logs
```bash
podman logs -f sliding-puzzle
```

### Stop the application
```bash
./stop.sh
```

Or manually:
```bash
podman stop sliding-puzzle
podman rm sliding-puzzle
```

### Restart after configuration changes
```bash
podman restart sliding-puzzle
```

### Rebuild after code changes
```bash
podman build -t sliding-puzzle:latest .
./stop.sh
./start.sh
```

## Development

### Running Locally (without container)

1. Install Go 1.21 or later
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Generate Templ templates:
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   templ generate
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

### Project Structure

```
puzzle/
├── main.go                    # Application entry point
├── config.yaml                # Configuration file
├── go.mod                     # Go dependencies
├── handlers/
│   ├── game.go               # Game page handler
│   └── completion.go         # Completion endpoint handler
├── models/
│   ├── config.go             # Configuration management
│   ├── database.go           # Database initialization
│   └── completion.go         # Completion model & logic
├── templates/
│   ├── layout.templ          # Base HTML layout
│   ├── game.templ            # Game page template
│   └── completion.templ      # Success/error templates
├── static/
│   ├── puzzle.js             # Puzzle game logic
│   └── style.css             # Styles
├── images/                    # PNG puzzle images
├── data/                      # SQLite database
├── Dockerfile                 # Container build
└── compose.yaml               # Container orchestration
```

## How It Works

1. **User visits the site**: A random PNG image from the `images/` directory is selected and sliced into a grid
2. **Puzzle gameplay**: User clicks tiles adjacent to the empty space to slide them
3. **Completion**: When solved, a modal prompts for an email address
4. **Code generation**: Server generates a unique discount code (format: `PUZZLE{%}-XXXXXX`)
5. **Storage**: Email and code are saved to SQLite database
6. **Display**: User receives their discount code

## Troubleshooting

### No images found error

**Problem**: Application shows "no PNG images found" error

**Solution**: Add at least one PNG file to the `images/` directory and restart:
```bash
cp your-image.png images/
podman-compose restart
```

### Port already in use

**Problem**: Port 8080 is already in use

**Solution**: Change the port in both `config.yaml` and `compose.yaml`:
```yaml
# config.yaml
server:
  port: 8081

# compose.yaml
ports:
  - "8081:8081"
```

### Database locked error

**Problem**: SQLite database is locked

**Solution**: Ensure only one instance is running:
```bash
podman-compose down
podman-compose up -d
```

### Container won't start

**Problem**: Container fails to start

**Solution**: Check logs for details:
```bash
podman-compose logs
```

## Security Notes

- This is a marketing tool intended for promotional use
- Email validation is basic; consider adding more robust validation for production
- No authentication is implemented; anyone can access the puzzle
- Discount codes are randomly generated and stored; implement validation in your e-commerce system
- Consider rate limiting if deploying publicly to prevent abuse

## License

This project is provided as-is for use by Repute Software.

## Support

For issues or questions, please contact the development team.

