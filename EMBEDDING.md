# Iframe Embedding Guide

Embed the Repute Puzzle on your website using a simple iframe. The puzzle includes parent window communication via postMessage and customizable styling.

## Quick Start

### Basic Embed

Add this iframe code to your website:

```html
<iframe 
    src="https://yoursite.com:8080/embed"
    width="600" 
    height="700"
    frameborder="0"
    style="border: none; border-radius: 10px;">
</iframe>
```

### With Custom Styling

Customize the primary color to match your brand:

```html
<iframe 
    src="https://yoursite.com:8080/embed?primaryColor=ff5722&lang=nl"
    width="600" 
    height="700"
    frameborder="0">
</iframe>
```

## URL Parameters

Customize the embedded puzzle with these URL parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `lang` | string | `en` | Language: `en` (English) or `nl` (Dutch) |
| `primaryColor` | string | `667eea` | Brand color in hex format (without #) |

### Examples

**English with red theme:**
```
https://yoursite.com:8080/embed?lang=en&primaryColor=e63946
```

**Dutch with blue theme:**
```
https://yoursite.com:8080/embed?lang=nl&primaryColor=4361ee
```

**Default (English, purple):**
```
https://yoursite.com:8080/embed
```

## Styling Guide

### Recommended Sizes

**Desktop:**
```html
<iframe src="..." width="600" height="700"></iframe>
```

**Mobile-responsive:**
```html
<style>
  .puzzle-embed {
    width: 100%;
    max-width: 600px;
    height: 700px;
    border: none;
    border-radius: 10px;
  }
  
  @media (max-width: 640px) {
    .puzzle-embed {
      height: 650px;
    }
  }
</style>

<iframe 
    class="puzzle-embed"
    src="https://yoursite.com:8080/embed">
</iframe>
```

### Brand Colors

The `primaryColor` parameter affects:
- Button backgrounds
- Gradient backgrounds
- Hover effects
- Stats display
- Modal headers

**Example colors:**
- Red: `e63946`
- Orange: `ff5722`
- Blue: `4361ee`
- Green: `06a77d`
- Purple: `667eea` (default)

## Parent Window Communication

The embedded puzzle sends events to your parent page using `postMessage`. Listen for these events to track user behavior.

### Event Structure

All events follow this format:

```javascript
{
  source: 'repute-puzzle',
  type: 'event.name',
  data: { ... },
  timestamp: 1697123456789
}
```

### Available Events

#### 1. `puzzle.loaded`
Fired when the puzzle initializes.

```javascript
{
  source: 'repute-puzzle',
  type: 'puzzle.loaded',
  data: {
    gridSize: 3,
    timeLimit: 90
  },
  timestamp: 1697123456789
}
```

#### 2. `puzzle.started`
Fired when the user makes their first move.

```javascript
{
  source: 'repute-puzzle',
  type: 'puzzle.started',
  data: {
    gridSize: 3
  },
  timestamp: 1697123456789
}
```

#### 3. `puzzle.completed`
Fired when the user solves the puzzle.

```javascript
{
  source: 'repute-puzzle',
  type: 'puzzle.completed',
  data: {
    moves: 45,
    time: 78,
    gridSize: 3
  },
  timestamp: 1697123456789
}
```

#### 4. `discount.generated`
Fired when a discount code is created for the user.

```javascript
{
  source: 'repute-puzzle',
  type: 'discount.generated',
  data: {
    code: 'PUZZLE15-ABC123',
    moves: 45,
    time: 78
  },
  timestamp: 1697123456789
}
```

#### 5. `puzzle.reset`
Fired when the user starts a new puzzle.

```javascript
{
  source: 'repute-puzzle',
  type: 'puzzle.reset',
  data: {
    gridSize: 3
  },
  timestamp: 1697123456789
}
```

### Example: Listening for Events

Add this JavaScript to your parent page:

```javascript
// Listen for messages from the puzzle iframe
window.addEventListener('message', function(event) {
  // Verify the message is from our puzzle
  if (event.data.source !== 'repute-puzzle') {
    return;
  }
  
  // Handle different event types
  switch(event.data.type) {
    case 'puzzle.loaded':
      console.log('Puzzle loaded!', event.data.data);
      // Track page view in analytics
      break;
      
    case 'puzzle.started':
      console.log('User started puzzle!');
      // Track engagement
      break;
      
    case 'puzzle.completed':
      console.log('Puzzle completed!', event.data.data);
      // Track conversion
      gtag('event', 'puzzle_completed', {
        moves: event.data.data.moves,
        time: event.data.data.time
      });
      break;
      
    case 'discount.generated':
      console.log('Discount code generated:', event.data.data.code);
      // Show success message on parent page
      document.getElementById('success-banner').style.display = 'block';
      // Track lead generation
      gtag('event', 'lead', {
        discount_code: event.data.data.code
      });
      break;
      
    case 'puzzle.reset':
      console.log('New puzzle started');
      break;
  }
});
```

### Example: Google Analytics Integration

Track puzzle interactions in Google Analytics:

```javascript
window.addEventListener('message', function(event) {
  if (event.data.source !== 'repute-puzzle') return;
  
  const { type, data } = event.data;
  
  // Google Analytics 4
  if (type === 'puzzle.completed') {
    gtag('event', 'puzzle_completed', {
      event_category: 'engagement',
      event_label: 'puzzle_game',
      value: data.moves,
      time_seconds: data.time
    });
  }
  
  if (type === 'discount.generated') {
    gtag('event', 'generate_lead', {
      event_category: 'conversion',
      event_label: 'discount_code',
      discount_code: data.code
    });
  }
});
```

### Example: Display Success Message

Show a message on your page when user gets a discount code:

```html
<div id="puzzle-container">
  <iframe src="https://yoursite.com:8080/embed" width="600" height="700"></iframe>
</div>

<div id="success-message" style="display: none;">
  <h3>🎉 Congratulations!</h3>
  <p>Check your email for your discount code!</p>
</div>

<script>
window.addEventListener('message', function(event) {
  if (event.data.source === 'repute-puzzle' && event.data.type === 'discount.generated') {
    // Hide puzzle, show success message
    document.getElementById('puzzle-container').style.display = 'none';
    document.getElementById('success-message').style.display = 'block';
    
    // Scroll to message
    document.getElementById('success-message').scrollIntoView({ 
      behavior: 'smooth' 
    });
  }
});
</script>
```

## Complete Example

Here's a full example page with responsive design and event tracking:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Win a Discount!</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        
        .puzzle-section {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        
        h1 {
            text-align: center;
            color: #333;
        }
        
        .puzzle-embed {
            width: 100%;
            max-width: 600px;
            height: 700px;
            border: none;
            border-radius: 10px;
            margin: 0 auto;
            display: block;
        }
        
        .stats {
            text-align: center;
            margin-top: 20px;
            color: #666;
        }
        
        @media (max-width: 640px) {
            .puzzle-embed {
                height: 650px;
            }
            .puzzle-section {
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="puzzle-section">
        <h1>🧩 Solve the Puzzle, Win 15% Off!</h1>
        <p style="text-align: center; color: #666;">
            Complete the sliding puzzle to receive your discount code
        </p>
        
        <iframe 
            class="puzzle-embed"
            src="https://yoursite.com:8080/embed?primaryColor=e63946&lang=en"
            title="Puzzle Game">
        </iframe>
        
        <div class="stats" id="stats"></div>
    </div>

    <script>
        // Track puzzle events
        window.addEventListener('message', function(event) {
            if (event.data.source !== 'repute-puzzle') return;
            
            const stats = document.getElementById('stats');
            
            switch(event.data.type) {
                case 'puzzle.loaded':
                    stats.textContent = '✨ Puzzle loaded! Start sliding tiles...';
                    break;
                    
                case 'puzzle.started':
                    stats.textContent = '🎮 Good luck!';
                    break;
                    
                case 'puzzle.completed':
                    stats.textContent = `🎉 Completed in ${event.data.data.moves} moves and ${event.data.data.time} seconds!`;
                    break;
                    
                case 'discount.generated':
                    stats.textContent = `✅ Code ${event.data.data.code} sent to your email!`;
                    
                    // Send to your analytics
                    if (typeof gtag !== 'undefined') {
                        gtag('event', 'discount_generated', {
                            code: event.data.data.code
                        });
                    }
                    break;
            }
        });
    </script>
</body>
</html>
```

## Security Considerations

### CORS and iframe Security

The puzzle app allows iframe embedding. If you need to restrict which domains can embed the puzzle, configure `X-Frame-Options` or `Content-Security-Policy` headers on your server.

### PostMessage Security

The puzzle sends messages to `*` (any origin). Your parent page should:

1. **Always verify the message source:**
   ```javascript
   if (event.data.source !== 'repute-puzzle') return;
   ```

2. **Validate data before using:**
   ```javascript
   if (typeof event.data.data.code === 'string') {
     // Safe to use
   }
   ```

3. **Never execute code from messages:**
   ```javascript
   // DON'T DO THIS:
   eval(event.data.code); // DANGEROUS!
   ```

## Troubleshooting

### Iframe Not Loading

**Problem:** Blank iframe or "refused to connect" error

**Solutions:**
1. Check the URL is correct
2. Ensure your server allows iframe embedding
3. Check browser console for errors
4. Try accessing the embed URL directly in a browser

### Colors Not Applying

**Problem:** Custom `primaryColor` not showing

**Solutions:**
1. Ensure color is in hex format without `#` (e.g., `ff5722` not `#ff5722`)
2. Check the color is 6 characters long
3. Clear browser cache
4. Try a different color to test

### Events Not Firing

**Problem:** Not receiving postMessage events

**Solutions:**
1. Verify your event listener is set up before the iframe loads
2. Check the message source is `'repute-puzzle'`
3. Open browser console and log all messages to debug:
   ```javascript
   window.addEventListener('message', (e) => console.log(e.data));
   ```

### Mobile Display Issues

**Problem:** Puzzle too small or cut off on mobile

**Solutions:**
1. Use responsive CSS (see examples above)
2. Reduce iframe height on mobile: `@media (max-width: 640px)`
3. Set iframe width to `100%` with max-width
4. Test on actual devices, not just browser dev tools

## Support

- **Documentation:** [README.md](README.md)
- **GitHub Issues:** [Report a bug](https://github.com/Repute-Software/Repute-Puzzle/issues)
- **Email Setup:** [EMAIL_SETUP.md](EMAIL_SETUP.md)
- **Security:** [SECURITY.md](SECURITY.md)

---

**Happy Embedding! 🎉**

