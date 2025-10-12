# Sliding Puzzle Marketing Tool

A containerized web application that serves as an interactive sliding puzzle game for marketing campaigns. Users complete the puzzle to receive a discount code, which is emailed to them automatically.

## Features

- 🧩 **Interactive Sliding Puzzle** - Configurable grid size (3x3, 4x4, 5x5, etc.)
- 🎨 **Custom Images** - Upload your own puzzle images
- 📧 **Email Integration** - Automatic discount code delivery via Resend
- 🌍 **Bilingual** - English and Dutch support
- ⏱️ **Time Tracking** - Optional time limits
- 🧪 **Testing Mode** - Highlight correct moves and auto-solve for testing
- 🐳 **Fully Containerized** - Easy deployment with Docker/Podman
- 🔐 **Database Storage** - SQLite stores completions and codes
- 🎯 **GitHub Actions CI/CD** - Automated builds and deployments

## Quick Start

### First Time Setup

1. **Clone or download this repository**

2. **Copy the example config**:
   ```bash
   cp config.example.yaml config.yaml
   ```

3. **Edit config.yaml**:
   - Set your Resend API key (see Email Setup below)
   - Customize puzzle settings (grid size, discount %, time limit)
   - Update sender name and email
   ```yaml
   puzzle:
     grid_size: 3          # 3x3 puzzle
     discount_percent: 15  # 15% discount
   
   email:
     api_key: "re_YOUR_KEY_HERE"
     from_email: "noreply@yourdomain.com"
     from_name: "Your Company"
   ```

4. **Add your puzzle images**:
   ```bash
   # Place PNG images in the images/ directory
   cp your-image.png images/
   ```

5. **Start the application**:
   ```bash
   ./start.sh
   ```

6. **Open in browser**:
   ```
   http://localhost:8080
   ```

## Configuration

### config.yaml

Main configuration file (not tracked in git):

```yaml
puzzle:
  grid_size: 3          # Puzzle grid size (3=3x3, 4=4x4, etc.)
  discount_percent: 15  # Discount percentage for code
  time_limit: 90        # Time limit in seconds (0 = no limit)
  testing_mode: true    # Show hints and auto-solve button
  scramble_moves: 25    # Number of scramble moves (more = harder)
  auto_solve_speed: 50  # Auto-solve animation speed (ms per move)

email:
  enabled: true                          # Enable/disable email sending
  provider: "resend"                     # Email provider
  api_key: "re_YOUR_API_KEY_HERE"       # From https://resend.com/api-keys
  from_email: "noreply@yourdomain.com"  # Verified sender email
  from_name: "Your Company"              # Sender display name
  subject_en: "Your Discount Code!"      # English subject
  subject_nl: "Jouw Kortingscode!"      # Dutch subject

database:
  path: "./data/puzzle.db"  # SQLite database location

server:
  port: 8080  # HTTP port

images:
  directory: "./images"  # Puzzle images directory
```

## Email Setup

This application uses [Resend](https://resend.com) for sending emails. Free tier includes 3,000 emails/month.

### Steps:

1. **Create Resend Account**:
   - Go to https://resend.com
   - Sign up (free)
   - Verify your email

2. **Get API Key**:
   - Go to https://resend.com/api-keys
   - Click "Create API Key"
   - Name: "Puzzle App"
   - Copy the key (starts with `re_...`)

3. **Update config.yaml**:
   ```yaml
   email:
     api_key: "re_your_actual_key_here"
   ```

4. **Verify Domain** (optional but recommended):
   - Go to https://resend.com/domains
   - Add your domain
   - Add DNS records (SPF, DKIM, DMARC)
   - Wait for verification
   - Or use `onboarding@resend.dev` for testing

See [EMAIL_SETUP.md](EMAIL_SETUP.md) for detailed instructions.

## Running Locally

### Using Helper Scripts (Recommended)

```bash
# Start
./start.sh

# Stop
./stop.sh
```

### Using Docker Compose Directly

```bash
# Start
docker compose up -d

# Stop
docker compose down

# View logs
docker compose logs -f
```

### Manual Build

```bash
# Build image
podman build -t sliding-puzzle:latest .

# Run container
podman run -d \
  -p 8080:8080 \
  -v ./images:/app/images:ro \
  -v ./data:/app/data \
  -v ./config.yaml:/app/config.yaml:ro \
  --name puzzle \
  sliding-puzzle:latest
```

## Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for complete production deployment guide.

### Quick Production Setup:

1. **On production server**:
   ```bash
   # Clone repo
   git clone https://github.com/Repute-Software/Repute-Puzzle.git
   cd Repute-Puzzle
   
   # Copy and configure
   cp config.example.yaml config.yaml
   nano config.yaml  # Add your settings
   
   # Add images
   cp your-images.png images/
   
   # Deploy
   docker compose -f compose.prod.yaml up -d
   ```

2. **Update**:
   ```bash
   docker compose -f compose.prod.yaml pull
   docker compose -f compose.prod.yaml up -d
   ```

## File Structure

```
puzzle/
├── config.example.yaml      # Example config (tracked in git)
├── config.yaml              # Your config (NOT in git)
├── compose.yaml             # Local development
├── compose.prod.yaml        # Production deployment
├── Dockerfile               # Container image
├── main.go                  # Application entry point
├── models/                  # Data models
│   ├── config.go           # Configuration
│   ├── database.go         # SQLite database
│   ├── discount.go         # Code generation
│   ├── email.go            # Email service
│   └── i18n.go             # Translations
├── handlers/                # HTTP handlers
│   ├── game.go             # Main game page
│   └── completion.go       # Puzzle completion
├── templates/               # HTML templates (Templ)
│   ├── layout.templ
│   ├── game.templ
│   └── completion.templ
├── static/                  # Static assets
│   ├── puzzle.js           # Game logic
│   └── style.css           # Styling
├── locales/                 # Translations
│   ├── en.yaml             # English
│   └── nl.yaml             # Dutch
├── images/                  # Puzzle images (you provide)
│   └── README.md
├── data/                    # Database (auto-created)
│   └── puzzle.db
└── docs/                    # Documentation
    ├── EMAIL_SETUP.md      # Email setup guide
    ├── DEPLOYMENT.md       # Production deployment
    └── SECURITY.md         # Security best practices
```

## Images

Place PNG images in the `images/` directory. The app will randomly select one for each game.

**Requirements**:
- Format: PNG
- Recommended size: 600x600px or larger
- Square images work best
- Multiple images supported (random selection)

## Database

Completions are stored in SQLite database at `./data/puzzle.db`.

**Schema**:
```sql
CREATE TABLE completions (
    id INTEGER PRIMARY KEY,
    email TEXT,
    discount_code TEXT UNIQUE,
    grid_size INTEGER,
    moves INTEGER,
    time_seconds INTEGER,
    completed_at DATETIME
);
```

## Testing Mode

Enable in `config.yaml`:
```yaml
puzzle:
  testing_mode: true
```

**Features**:
- Green highlight shows correct next move
- "Auto-Solve" button to complete puzzle
- Useful for testing email flow

**Remember to disable in production!**

## Languages

Switch language with URL parameter:
- English: `http://localhost:8080?lang=en`
- Dutch: `http://localhost:8080?lang=nl`

Add more languages by:
1. Create `locales/xx.yaml` (copy from `en.yaml`)
2. Translate all strings
3. Users access with `?lang=xx`

## Development

### Prerequisites

- Go 1.23+
- Templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`
- Docker or Podman

### Build Steps

```bash
# Generate templates
templ generate

# Build binary
go build -o puzzle .

# Run locally
./puzzle
```

### Hot Reload (Development)

```bash
# Terminal 1: Watch and rebuild templates
templ generate --watch

# Terminal 2: Watch and rebuild Go
go run .
```

## Monitoring

### View Logs

```bash
# Local
docker compose logs -f

# Production
docker compose -f compose.prod.yaml logs -f

# Filter for email
docker compose logs | grep email
```

### Check Status

```bash
docker compose ps
docker compose -f compose.prod.yaml ps
```

## Backup

Important data to backup:
- `data/puzzle.db` - All completions and codes
- `config.yaml` - Your configuration (contains API key!)
- `images/` - Your puzzle images

```bash
# Create backup
tar -czf puzzle-backup-$(date +%Y%m%d).tar.gz \
  data/puzzle.db config.yaml images/

# Restore
tar -xzf puzzle-backup-20241012.tar.gz
```

## Troubleshooting

### Email Not Sending

1. **Check API key**:
   ```bash
   grep api_key config.yaml
   ```

2. **Check logs**:
   ```bash
   docker compose logs | grep -i email
   docker compose logs | grep -i error
   ```

3. **Verify domain** at https://resend.com/domains

4. **Test with**:
   - `onboarding@resend.dev` (works immediately)
   - Your verified domain

### Port Already in Use

```bash
# Change port in config.yaml
server:
  port: 8081  # Use different port

# Or in compose
ports:
  - "8081:8080"
```

### Database Locked

```bash
# Stop all containers
docker compose down

# Remove database lock
rm data/puzzle.db-shm data/puzzle.db-wal

# Restart
docker compose up -d
```

## Support

- **Documentation**: See `/docs` folder
- **Issues**: https://github.com/Repute-Software/Repute-Puzzle/issues
- **Email**: See [EMAIL_SETUP.md](EMAIL_SETUP.md)
- **Security**: See [SECURITY.md](SECURITY.md)

## License

[Your License Here]

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

**Happy Puzzling! 🧩**
