# Email Integration Setup Guide

This guide walks you through setting up Resend email integration to automatically send discount codes to users.

## Overview

The puzzle application now sends beautiful HTML emails with discount codes to users after they complete the puzzle. Email sending happens in the background so the user experience remains fast.

## Prerequisites

### Step 1: Create Resend Account

1. Go to https://resend.com
2. Click "Start Building" or "Sign Up"
3. Sign up with GitHub (easiest) or email
4. Verify your email address

**Why Resend?**
- ✅ Free tier: 3,000 emails/month, 100/day
- ✅ Simple API (just 1 HTTP call)
- ✅ Reliable delivery
- ✅ Great developer experience

### Step 2: Get API Key

1. Go to https://resend.com/api-keys
2. Click "Create API Key"
3. Name it: "Puzzle App Production"
4. Permissions: "Sending access"
5. Click "Add"
6. **Copy the API key** (starts with `re_...`)
7. **Save it securely** - you won't see it again!

Example API key format: `re_123abc456def789ghi`

### Step 3: Configure Sender Email

You have two options:

#### Option A: Use Resend's Domain (Quick Start) ⚡

**Best for**: Testing, quick setup

- Sender email: `onboarding@resend.dev`
- No DNS setup needed
- Works immediately
- Already configured in `config.yaml`

#### Option B: Use Your Own Domain (Professional) 🏢

**Best for**: Production, branded emails

1. Go to https://resend.com/domains
2. Click "Add Domain"
3. Enter your domain: `yourdomain.com`
4. Add these DNS records (provided by Resend):
   - **SPF** (TXT record)
   - **DKIM** (TXT record)
   - **DMARC** (TXT record)
5. Wait for verification (5-30 minutes)
6. Update `config.yaml`:
   ```yaml
   email:
     from_email: "noreply@yourdomain.com"
   ```

## Configuration

### Local Development

1. **Copy env.example to .env**:
   ```bash
   cp env.example .env
   ```

2. **Edit .env and add your API key**:
   ```bash
   RESEND_API_KEY=re_your_actual_api_key_here
   ```

3. **Update config.yaml** (optional):
   ```yaml
   email:
     enabled: true                          # Enable email sending
     provider: "resend"
     api_key: ""                            # Leave empty (uses env var)
     from_email: "onboarding@resend.dev"   # Or your domain
     from_name: "Puzzle Game"               # Customize this
     subject_en: "Your Discount Code!"
     subject_nl: "Jouw Kortingscode!"
   ```

4. **Test locally**:
   ```bash
   # Export API key
   export RESEND_API_KEY=re_your_actual_key_here
   
   # Rebuild and run
   ./stop.sh
   podman build -t sliding-puzzle:latest .
   ./start.sh
   
   # Complete a puzzle with your real email
   # Check your inbox!
   ```

### Production Deployment

1. **On production server, create .env file**:
   ```bash
   cd /path/to/puzzle
   
   # Create .env file (never commit to git!)
   echo "RESEND_API_KEY=re_your_actual_api_key_here" > .env
   
   # Secure it
   chmod 600 .env
   ```

2. **Update config.yaml**:
   ```yaml
   email:
     enabled: true
     from_email: "noreply@yourdomain.com"  # Your verified domain
     from_name: "Your Company"
   ```

3. **Deploy**:
   ```bash
   # Pull latest image
   docker compose -f compose.prod.yaml pull
   
   # Start with environment file
   docker compose -f compose.prod.yaml --env-file .env up -d
   
   # Check logs
   docker compose -f compose.prod.yaml logs -f
   ```

## Email Features

### Beautiful HTML Template

The email includes:
- 🎨 Gradient header matching your app design
- 💎 Prominent discount code display
- 🌍 Bilingual support (English/Dutch)
- 📱 Mobile-responsive design
- ✉️ Professional formatting

### Example Email (English)

```
╔═══════════════════════════════════════╗
║                                       ║
║     🎉 Congratulations!              ║
║                                       ║
╠═══════════════════════════════════════╣
║                                       ║
║  You've completed the puzzle!         ║
║  Here's your discount code:           ║
║                                       ║
║  ┌───────────────────────────────┐   ║
║  │   PUZZLE15-ABC123            │   ║
║  └───────────────────────────────┘   ║
║                                       ║
║  Save 15% on your next purchase!      ║
║                                       ║
║  Use this code at checkout to         ║
║  receive your discount.               ║
║                                       ║
║        [Play Again]                   ║
║                                       ║
╚═══════════════════════════════════════╝
```

### Language Support

Email language is automatically selected based on the user's interface language:
- English: `?lang=en` or default
- Dutch: `?lang=nl`

## Testing

### Test Without Playing

Create a test script:

```bash
#!/bin/bash
# test-email.sh

# Test with your email
EMAIL="your@email.com"
API_KEY="re_your_key_here"

# Export API key
export RESEND_API_KEY=$API_KEY

# Start local server
./start.sh

# Wait for startup
sleep 3

# Submit completion (simulates puzzle completion)
curl -X POST http://localhost:8080/complete \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "email=$EMAIL&moves=25&time=120&lang=en"

echo "Check your email at $EMAIL"
```

### Verify Email Delivery

1. **Check application logs**:
   ```bash
   docker compose -f compose.prod.yaml logs | grep email
   
   # Should see:
   # Successfully sent discount code email to user@example.com
   ```

2. **Check Resend dashboard**:
   - Go to https://resend.com/emails
   - See all sent emails
   - View delivery status
   - Check email content

3. **Check user's inbox**:
   - Email should arrive within seconds
   - Check spam folder if not visible

## Troubleshooting

### Email Not Sending

**Problem**: No email received

**Solutions**:
1. Check API key is set:
   ```bash
   docker compose -f compose.prod.yaml exec puzzle env | grep RESEND
   ```

2. Check logs for errors:
   ```bash
   docker compose -f compose.prod.yaml logs | grep -i email
   docker compose -f compose.prod.yaml logs | grep -i error
   ```

3. Verify email is enabled in config.yaml:
   ```yaml
   email:
     enabled: true
   ```

4. Test API key manually:
   ```bash
   curl -X POST 'https://api.resend.com/emails' \
     -H 'Authorization: Bearer re_your_key_here' \
     -H 'Content-Type: application/json' \
     -d '{
       "from": "onboarding@resend.dev",
       "to": ["your@email.com"],
       "subject": "Test",
       "html": "<p>Test email</p>"
     }'
   ```

### Wrong Sender Email

**Problem**: Emails sent from wrong address

**Solution**: Update config.yaml:
```yaml
email:
  from_email: "correct@yourdomain.com"
  from_name: "Your Company"
```

Restart:
```bash
docker compose -f compose.prod.yaml restart
```

### API Rate Limits

**Free tier limits**:
- 3,000 emails/month
- 100 emails/day

**If you exceed limits**:
1. Upgrade Resend plan
2. Or disable email temporarily:
   ```yaml
   email:
     enabled: false
   ```

### Domain Not Verified

**Problem**: Using your domain before DNS verification

**Solution**:
1. Wait for DNS propagation (up to 30 minutes)
2. Check verification status at https://resend.com/domains
3. Use `onboarding@resend.dev` in the meantime

## Configuration Reference

### config.yaml

```yaml
email:
  enabled: true                    # true/false - enable email sending
  provider: "resend"               # Email service provider
  api_key: ""                      # Leave empty (set via env var)
  from_email: "sender@domain.com"  # Verified sender email
  from_name: "Your Company"        # Sender display name
  subject_en: "Your Discount!"     # English subject line
  subject_nl: "Jouw Korting!"     # Dutch subject line
```

### Environment Variables

```bash
# Required for email sending
RESEND_API_KEY=re_your_api_key_here
```

## Security Best Practices

1. **Never commit API keys to git**:
   - .env is in .gitignore
   - env.example is safe (no real keys)

2. **Secure .env file**:
   ```bash
   chmod 600 .env  # Only owner can read
   ```

3. **Use environment variables in production**:
   - Don't put API key in config.yaml
   - Use RESEND_API_KEY environment variable

4. **Rotate API keys periodically**:
   - Create new key at https://resend.com/api-keys
   - Update .env
   - Delete old key

5. **Monitor usage**:
   - Check https://resend.com/emails
   - Watch for suspicious activity

## Advanced Features

### Disable Email Temporarily

Set in config.yaml:
```yaml
email:
  enabled: false  # Emails disabled, app still works
```

### Custom Email Templates

Edit `models/email.go`, function `generateEmailHTML()`:
- Modify HTML structure
- Change colors/styling
- Add your logo
- Customize text

### Add More Languages

1. Add translation in `models/email.go`:
   ```go
   if lang == "es" {
       data["Subject"] = "Tu Código de Descuento"
       // ... more Spanish translations
   }
   ```

2. Add subject to config.yaml:
   ```yaml
   email:
     subject_es: "¡Tu Código de Descuento!"
   ```

## Cost Estimate

**Free tier** (perfect for small to medium use):
- 3,000 emails/month = ~100/day
- Enough for 100 daily puzzle completions

**Paid tier** (if you need more):
- $20/month = 50,000 emails
- $0.0004 per additional email

**Example costs**:
- 100 completions/day × 30 days = 3,000 emails = **FREE**
- 500 completions/day × 30 days = 15,000 emails = **$20/month**
- 1,000 completions/day × 30 days = 30,000 emails = **$20/month**

## Support

- **Resend Docs**: https://resend.com/docs
- **Resend API**: https://resend.com/docs/api-reference/emails/send-email
- **Resend Status**: https://status.resend.com
- **Resend Support**: support@resend.com

---

**Status: Email Integration Complete! ✅**

Users will now automatically receive their discount codes via email!

