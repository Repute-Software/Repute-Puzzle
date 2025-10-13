# Testing Guide - Multi-Tenant Puzzle Platform

## Current Status

**Branch:** `feature/multi-tenant`

**Completed:**
- ✅ Phase 1: Database schema & models (companies, users, sessions, puzzles)
- ✅ Phase 2: Authentication system (login, signup, logout)
- ✅ Phase 3 (Partial): Admin templates created (but handlers not yet wired up)

**What Works:**
- Database migrations
- User authentication with bcrypt
- Session management  
- Beautiful login/signup templates
- Admin dashboard templates (visual only, no backend yet)

**What Doesn't Work Yet:**
- Admin dashboard (needs handlers in `handlers/admin.go`)
- Puzzle CRUD operations (needs admin routes in `main.go`)
- Image upload
- New game routes (`/play/:company/:puzzle`)

---

## Manual Testing Steps

### 1. Start the Server

```bash
cd /home/admindt/Projecten/repute-software/puzzle
./puzzle
```

Server should start on `http://localhost:8080`

### 2. Test Health Check

```bash
curl http://localhost:8080/health
```

Expected: `OK`

### 3. Test Signup Flow

**Option A: Browser**
1. Open `http://localhost:8080/signup` in your browser
2. Fill in:
   - Company Name: "Test Company"
   - Email: "admin@test.com"  
   - Password: "Test1234"
   - Confirm Password: "Test1234"
3. Click "Create Account"
4. Should redirect to `/admin` (shows placeholder text for now)

**Option B: Command Line**
```bash
curl -v -X POST http://localhost:8080/signup \
  -d "company_name=Test Company" \
  -d "email=admin@test.com" \
  -d "password=Test1234" \
  -d "confirm_password=Test1234" \
  -c cookies.txt
```

Expected: 303 redirect to `/admin`, session cookie set

### 4. Test Login Flow

**Option A: Browser**
1. Open `http://localhost:8080/login`
2. Enter:
   - Email: "admin@test.com"
   - Password: "Test1234"
3. Click "Sign In"
4. Should redirect to `/admin`

**Option B: Command Line**
```bash
curl -v -X POST http://localhost:8080/login \
  -d "email=admin@test.com" \
  -d "password=Test1234" \
  -c cookies.txt
```

### 5. Test Protected Route

```bash
# Without auth (should redirect)
curl -v http://localhost:8080/admin

# With auth (should show welcome)
curl -v http://localhost:8080/admin -b cookies.txt
```

### 6. Test Logout

```bash
curl -v http://localhost:8080/logout -b cookies.txt
```

Expected: 303 redirect to `/login`, session cleared

### 7. Check Database

```bash
sqlite3 ./data/puzzle.db

# View created companies
SELECT * FROM companies;

# View created users
SELECT * FROM users;

# View sessions
SELECT * FROM sessions;

# Exit
.exit
```

---

## Known Issues

1. **Admin Dashboard Not Functional:**
   - Templates exist but no handlers
   - `/admin` shows placeholder text
   - Need to implement `handlers/admin.go`

2. **Old Routes Still Active:**
   - `/` still shows old game page
   - `/embed` still shows old embed page
   - These need to be updated to use puzzle slugs

3. **No Puzzles Yet:**
   - Can't create puzzles until admin handlers are done
   - Completion flow still uses hardcoded `puzzle_id = 1`

---

## Next Implementation Steps

1. **Create `handlers/admin.go`** (~200 lines)
   - Dashboard (list puzzles)
   - Create puzzle (with image upload)
   - Edit puzzle
   - Delete puzzle
   - View puzzle details

2. **Wire up admin routes in `main.go`** (~50 lines)
   - `/admin` → dashboard
   - `/admin/puzzles/new` → create form
   - `/admin/puzzles/:id` → detail view
   - `/admin/puzzles/:id/edit` → edit form
   - `/admin/puzzles/:id/delete` → delete action

3. **Update game routes** (~100 lines)
   - `/play/:company/:puzzle` → game page
   - `/embed/:company/:puzzle` → embed page
   - `/complete/:company/:puzzle` → completion

4. **Create default puzzle migration**
   - Seed database with default company and puzzle
   - Migrate existing completions

---

## Test Data

After signup, you should have:
- 1 company in `companies` table
- 1 admin user in `users` table
- 1 active session in `sessions` table (expires in 7 days)

To reset:
```bash
rm -f data/puzzle.db data/puzzle.db-*
# Restart server to recreate fresh database
```

---

## Visual Test (Browser Recommended)

1. **Login Page**: Clean, modern UI with purple gradient
2. **Signup Page**: Company creation form
3. **Admin Area**: Shows user email and company name in header
4. **Templates Ready**: Dashboard, form, and detail templates exist but need backend

---

## Architecture Verification

✅ **Sessions work** - Cookie-based auth with HttpOnly, Secure flags  
✅ **Password hashing** - bcrypt with cost 12  
✅ **Company isolation** - Each user belongs to a company  
✅ **Database schema** - All tables created with proper indexes  
✅ **Templates** - Login, signup, and admin UI all compiled  

❌ **Admin CRUD** - Not yet implemented  
❌ **Image upload** - Not yet implemented  
❌ **Multi-puzzle routes** - Not yet implemented  

---

## Production Readiness Checklist

- [ ] Admin handlers implemented
- [ ] Image upload with validation
- [ ] Puzzle CRUD operations
- [ ] New game routes with slug lookup
- [ ] Default puzzle migration
- [ ] CSRF protection on forms
- [ ] Rate limiting on login
- [ ] Session cleanup cron job
- [ ] Error logging to file
- [ ] Deploy to VPS
- [ ] Test embed in real website

Current estimate: **~60% complete**
