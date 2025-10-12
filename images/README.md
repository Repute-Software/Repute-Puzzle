# Puzzle Images Directory

Place your PNG puzzle images in this directory. The application will randomly select one image per game session.

## Requirements

- **Format**: PNG files only (*.png)
- **Size**: Minimum 600x600 pixels recommended
- **Aspect Ratio**: Square images (1:1) work best
- **Number**: At least 1 image required
- **File Size**: Keep under 2MB for faster loading

## Recommendations

### Image Selection
- Use high-quality, clear images
- Avoid images with too much detail (harder to solve)
- Use images related to your brand or products
- Consider seasonal or campaign-specific images

### Image Preparation
- **Resize**: 800x800 or 1000x1000 pixels is optimal
- **Optimize**: Use PNG compression to reduce file size
- **Test**: Try the puzzle yourself to ensure solvability

### Tools for Optimization
```bash
# Using ImageMagick to resize and optimize
convert input.jpg -resize 800x800 -quality 95 puzzle-image.png

# Using pngquant to compress
pngquant --quality=85-95 puzzle-image.png
```

## Adding Images

### Local Development
```bash
# Copy images directly
cp /path/to/your/image.png /home/admindt/Projecten/repute-software/puzzle/images/

# Restart container to use new images
./stop.sh && ./start.sh
```

### Production
```bash
# Copy images to production server
scp your-image.png user@server:/path/to/puzzle/images/

# No restart needed - changes are applied immediately
```

## Image Ideas

- **Product Photos**: Your best-selling products
- **Brand Logo**: Company logo or mascot
- **Seasonal**: Holiday-themed images
- **Events**: Conference or event photos
- **Team**: Group photos (if not too complex)
- **Artwork**: Custom illustrations or designs

## Testing Tips

When adding new images:

1. Set `testing_mode: true` in config.yaml
2. Set `scramble_moves: 10` for quick testing
3. Test the puzzle to ensure:
   - Image is clear when divided
   - Puzzle is fun to solve
   - Image quality is good
4. Set `testing_mode: false` before production use

## Current Images

List your images here for documentation:
- *Add your image filenames here*

---

**Note**: The application requires at least one PNG file in this directory to function.
