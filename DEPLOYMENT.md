# Production Deployment Guide

## Prerequisites

- Server with Docker/Podman installed
- GitHub account with access to the repository
- Domain name (optional, for public access)

## Quick Deployment

### 1. Clone Repository
```bash
git clone https://github.com/USERNAME/puzzle.git
cd puzzle
```

### 2. Configure
```bash
# Edit config.yaml
vim config.yaml

# Set production values:
puzzle:
  grid_size: 3
  discount_percent: 15
  time_limit: 300
  testing_mode: false  # IMPORTANT: Disable for production
  scramble_moves: 50
  auto_solve_speed: 50
```

### 3. Add Puzzle Images
```bash
# Copy your puzzle images
cp /path/to/your-images/*.png images/
```

### 4. Update compose.prod.yaml
```bash
# Replace USERNAME with your actual GitHub username
vim compose.prod.yaml
```

### 5. Deploy
```bash
# Using Docker
docker compose -f compose.prod.yaml up -d

# OR using Podman
podman-compose -f compose.prod.yaml up -d
```

### 6. Verify
```bash
curl http://localhost:8080
```

## Image Tags

- `latest` - Latest build from main branch
- `v1.0.0` - Specific version (use for production)
- `v1.0` - Minor version (auto-updates patch versions)
- `v1` - Major version (auto-updates minor/patch)

## Updating

### Update to Latest
```bash
docker compose -f compose.prod.yaml pull
docker compose -f compose.prod.yaml up -d
```

### Update to Specific Version
Edit `compose.prod.yaml`:
```yaml
image: ghcr.io/USERNAME/puzzle:v1.2.0
```
Then:
```bash
docker compose -f compose.prod.yaml pull
docker compose -f compose.prod.yaml up -d
```

## Monitoring

### View Logs
```bash
docker compose -f compose.prod.yaml logs -f
```

### Check Status
```bash
docker compose -f compose.prod.yaml ps
```

### View Database
```bash
docker exec -it sliding-puzzle sqlite3 /app/data/puzzle.db
sqlite> SELECT * FROM completions ORDER BY created_at DESC LIMIT 10;
sqlite> .quit
```

### Check Disk Usage
```bash
docker compose -f compose.prod.yaml exec puzzle du -sh /app/data
```

## Backup

### Backup Database
```bash
# Simple backup
cp data/puzzle.db data/puzzle.db.backup

# Backup with timestamp
cp data/puzzle.db data/puzzle_$(date +%Y%m%d_%H%M%S).db
```

### Automated Backup Script
Create `backup.sh`:
```bash
#!/bin/bash
# Backup puzzle database with rotation

BACKUP_DIR="backups"
mkdir -p "$BACKUP_DIR"

# Create backup with timestamp
DATE=$(date +%Y%m%d_%H%M%S)
cp data/puzzle.db "$BACKUP_DIR/puzzle_${DATE}.db"

# Keep only last 30 days
find "$BACKUP_DIR" -name "puzzle_*.db" -mtime +30 -delete

echo "Backup completed: puzzle_${DATE}.db"
```

Make it executable and add to cron:
```bash
chmod +x backup.sh

# Add to crontab (daily at 2 AM)
crontab -e
# Add: 0 2 * * * cd /path/to/puzzle && ./backup.sh
```

## Troubleshooting

### Container Won't Start
```bash
# Check logs
docker compose -f compose.prod.yaml logs

# Check if port is already in use
sudo netstat -tulpn | grep 8080

# Check container status
docker compose -f compose.prod.yaml ps -a
```

### Port Already in Use
Edit `compose.prod.yaml`:
```yaml
ports:
  - "8081:8080"  # Use different external port
```

### No Images Found
```bash
# Check images directory
ls -la images/*.png

# Ensure at least one PNG image exists
# Add a default image if needed
```

### Permission Issues
```bash
# Fix data directory permissions
sudo chown -R 1000:1000 data/

# Fix images directory permissions
sudo chmod 755 images/
```

### Database Locked
```bash
# Stop the container first
docker compose -f compose.prod.yaml down

# Then perform database operations
sqlite3 data/puzzle.db "SELECT * FROM completions;"

# Restart
docker compose -f compose.prod.yaml up -d
```

## Security Considerations

### Firewall
```bash
# Allow only specific IPs (replace with your IP range)
sudo ufw allow from 192.168.1.0/24 to any port 8080

# Or use a reverse proxy (recommended for production)
```

### Reverse Proxy with Nginx

Example nginx configuration:
```nginx
server {
    listen 80;
    server_name puzzle.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### SSL/TLS with Let's Encrypt
```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d puzzle.yourdomain.com
```

## Performance Optimization

### Resource Limits
Edit `compose.prod.yaml` to add resource constraints:
```yaml
services:
  puzzle:
    # ... existing config ...
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
```

### Log Rotation
The compose file already includes log rotation:
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## Multi-Language Support

The application supports multiple languages via URL parameter:
- English (default): `http://yourserver:8080`
- Dutch: `http://yourserver:8080?lang=nl`

Share language-specific links in your marketing campaigns.

## Maintenance

### Update Application Code
```bash
# Pull latest code
git pull origin main

# Pull new image
docker compose -f compose.prod.yaml pull

# Restart with new image
docker compose -f compose.prod.yaml up -d
```

### Clean Up Old Images
```bash
# Remove unused images
docker image prune -a

# Or specifically
docker rmi ghcr.io/USERNAME/puzzle:old-tag
```

### Database Maintenance
```bash
# Vacuum database (optimize)
docker compose -f compose.prod.yaml exec puzzle sqlite3 /app/data/puzzle.db "VACUUM;"

# Check database integrity
docker compose -f compose.prod.yaml exec puzzle sqlite3 /app/data/puzzle.db "PRAGMA integrity_check;"
```

## Scaling

For high traffic, consider:
1. **Load Balancer**: Multiple instances behind nginx/HAProxy
2. **Shared Database**: Use PostgreSQL instead of SQLite
3. **CDN**: Serve static assets via CDN
4. **Caching**: Add Redis for session management

## Support

For issues:
1. Check logs: `docker compose -f compose.prod.yaml logs -f`
2. Verify configuration: `cat config.yaml`
3. Check GitHub repository for updates
4. Review GitHub Actions for build failures

---

**Production Deployment Checklist:**
- [ ] Config file reviewed (testing_mode: false)
- [ ] Images added to images/ directory
- [ ] compose.prod.yaml updated with correct username
- [ ] Firewall configured
- [ ] Backups configured
- [ ] Monitoring/logs reviewed
- [ ] SSL/TLS configured (if public)
- [ ] Resource limits set
- [ ] Database backed up regularly

