# Security Best Practices

## API Key Management

### ⚠️ IMPORTANT: Never Commit API Keys!

Your Resend API key should **NEVER** be in `config.yaml` or any file that's committed to git.

### ✅ Correct Way: Use Environment Variables

#### Local Development

1. **Create `.env` file** (already in `.gitignore`):
   ```bash
   echo "RESEND_API_KEY=re_your_key_here" > .env
   ```

2. **Load environment variables when running**:
   
   **Option A: Using compose (recommended)**
   ```bash
   # Compose automatically loads .env file
   ./start.sh
   ```
   
   **Option B: Manual export**
   ```bash
   export RESEND_API_KEY=re_your_key_here
   ./puzzle
   ```

3. **Keep `config.yaml` clean**:
   ```yaml
   email:
     api_key: ""  # Leave empty!
   ```

#### Production Deployment

1. **On production server, create `.env`**:
   ```bash
   cd /path/to/puzzle
   echo "RESEND_API_KEY=re_your_production_key" > .env
   chmod 600 .env  # Secure it
   ```

2. **Start with environment file**:
   ```bash
   docker compose -f compose.prod.yaml --env-file .env up -d
   ```

### How It Works

The application automatically loads the API key from the environment variable:

```go
// In main.go
if apiKey := os.Getenv("RESEND_API_KEY"); apiKey != "" {
    config.Email.APIKey = apiKey
}
```

Priority order:
1. `RESEND_API_KEY` environment variable (highest priority)
2. `config.yaml` api_key field (should be empty)

### Files in Git vs Local Only

**✅ Safe to Commit (in git):**
- `config.yaml` (with `api_key: ""`)
- `env.example` (template only)
- `compose.prod.yaml`
- `.gitignore`

**❌ Never Commit (local only):**
- `.env` (contains real API key)
- `config.local.yaml` (if you create one)
- Any file with "re_" API keys

### Verifying Security

**Check what will be committed:**
```bash
git status
git diff config.yaml
```

**Make sure `.env` is ignored:**
```bash
git check-ignore .env
# Should output: .env

cat .gitignore | grep .env
# Should see: .env
```

**Check for accidental commits:**
```bash
# Search entire git history for API keys (if you suspect you committed one)
git log -S "re_" --all --oneline
```

### If You Accidentally Committed an API Key

**DON'T PANIC - but act quickly:**

1. **Revoke the compromised key immediately**:
   - Go to https://resend.com/api-keys
   - Delete the exposed API key
   - Create a new one

2. **Remove from git history** (if recently committed):
   ```bash
   # If it's the last commit
   git reset HEAD~1
   git add config.yaml  # Add the fixed version
   git commit -m "Remove API key from config"
   ```

3. **If already pushed to GitHub**:
   ```bash
   # Remove from git history (use with caution!)
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch config.yaml" \
     --prune-empty --tag-name-filter cat -- --all
   
   # Force push (warning: affects all users!)
   git push origin --force --all
   ```

4. **Better approach for pushed commits**:
   - Just revoke the key at https://resend.com/api-keys
   - Create a new one
   - Fix config.yaml and commit properly
   - The old key is now useless anyway

### Domain Verification

Your email is configured to send from `noreply@repute-software.com`. 

**To use this domain:**

1. Go to https://resend.com/domains
2. Add domain: `repute-software.com`
3. Add these DNS records to your domain registrar:

   **SPF Record** (TXT):
   ```
   Name: @
   Value: v=spf1 include:_spf.resend.com ~all
   ```

   **DKIM Record** (TXT):
   ```
   Name: resend._domainkey
   Value: [Provided by Resend - copy from dashboard]
   ```

   **DMARC Record** (TXT):
   ```
   Name: _dmarc
   Value: v=DMARC1; p=none; rua=mailto:dmarc@repute-software.com
   ```

4. Wait 5-30 minutes for DNS propagation
5. Verify in Resend dashboard

**Until domain is verified:**
- Emails will fail to send
- Check logs: `docker compose logs | grep email`
- Alternative: Use `onboarding@resend.dev` (works immediately)

### Testing Email Security

**Verify environment variable is loaded:**
```bash
# Local
export RESEND_API_KEY=re_test_key
./puzzle
# Check logs for: "Email API key loaded from environment variable"

# Production
docker compose -f compose.prod.yaml exec puzzle env | grep RESEND_API_KEY
# Should show: RESEND_API_KEY=re_your_key
```

**Test without exposing key:**
```bash
# Good: Key in .env file
docker compose up

# Bad: Key in command (visible in process list!)
RESEND_API_KEY=re_key docker compose up
```

### Rotating API Keys

**Best practice: Rotate every 3-6 months**

1. Create new API key at https://resend.com/api-keys
2. Update `.env` file locally and on production
3. Restart application
4. Test email sending
5. Delete old API key from Resend dashboard

### Multiple Environments

**Development / Staging / Production:**

```bash
# .env.dev
RESEND_API_KEY=re_dev_key

# .env.staging  
RESEND_API_KEY=re_staging_key

# .env.prod
RESEND_API_KEY=re_prod_key
```

Use the correct file:
```bash
docker compose --env-file .env.dev up
docker compose --env-file .env.prod up
```

### Security Checklist

- [ ] API key is NOT in `config.yaml`
- [ ] `.env` file exists and contains API key
- [ ] `.env` is in `.gitignore`
- [ ] `config.yaml` has `api_key: ""`
- [ ] Tested that email works with env var
- [ ] Domain is verified in Resend (or using resend.dev)
- [ ] `.env` has `chmod 600` permissions
- [ ] Production `.env` is created on server

### Monitoring

**Check for security issues:**

```bash
# Check if API key is in config
grep "re_" config.yaml
# Should return nothing (or just comments)

# Check if .env is ignored
git ls-files | grep .env
# Should return nothing

# View what would be committed
git add -A
git status
# Should NOT see .env file
```

---

**Your API key is now secure!** ✅

The key is in `.env` (not committed to git) and loaded via environment variable.

