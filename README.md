# LinkedIn Automation - Educational Proof of Concept

> ⚠️ **LEGAL DISCLAIMER**: This project is an **educational proof-of-concept** created strictly for technical evaluation and learning purposes. It demonstrates advanced browser automation, anti-detection techniques, and Go programming best practices. **DO NOT use this for actual LinkedIn automation** as it may violate LinkedIn's Terms of Service. The author assumes no responsibility for misuse.

## 📋 Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Stealth Techniques](#stealth-techniques)
- [Features](#features)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Database Schema](#database-schema)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Development](#development)

## 🎯 Overview

This project demonstrates a sophisticated LinkedIn automation system built in Go, showcasing:

- **Advanced Browser Automation** using Rod (Chrome DevTools Protocol)
- **11+ Stealth Techniques** to simulate human behavior
- **Clean Architecture** with modular, testable code
- **State Persistence** with SQLite for resume capability
- **Rate Limiting** and scheduling for realistic operation
- **Comprehensive Logging** for debugging and monitoring

### Key Capabilities

1. **Authentication**: Login with cookie persistence and session reuse
2. **Search**: Find profiles by job title, location, and keywords
3. **Connect**: Send personalized connection requests with notes
4. **Message**: Follow up with accepted connections
5. **Scheduling**: Operate only during business hours
6. **Stealth**: Evade detection with human-like behavior

## 🛠 Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go 1.21+ | Core application |
| Browser Automation | Rod | Chrome DevTools Protocol |
| Stealth | go-rod/stealth | Anti-detection |
| Database | SQLite3 | State persistence |
| Config | YAML + .env | Configuration management |
| Logging | Logrus | Structured logging |

## 🏗 Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                         Main Loop                            │
│  (Scheduling, Rate Limiting, Graceful Shutdown)             │
└────────────┬────────────────────────────────────────────────┘
             │
    ┌────────┴────────┐
    │                 │
┌───▼────┐      ┌────▼─────┐      ┌──────────┐
│  Auth  │      │  Search  │      │ Connect  │
│Service │      │ Service  │      │ Service  │
└───┬────┘      └────┬─────┘      └────┬─────┘
    │                │                  │
    └────────┬───────┴──────────────────┘
             │
    ┌────────▼─────────┐
    │  Browser Context │
    │   + Stealth      │
    └────────┬─────────┘
             │
    ┌────────▼─────────┐
    │   Rod Browser    │
    │  (Chromium CDP)  │
    └──────────────────┘
```

### Package Responsibilities

- **cmd/main.go**: Application entry point, orchestration
- **internal/auth**: Authentication and session management
- **internal/browser**: Browser initialization with stealth
- **internal/config**: Configuration loading and validation
- **internal/connect**: Connection request handling
- **internal/logger**: Structured logging
- **internal/message**: Messaging system
- **internal/scheduler**: Activity scheduling
- **internal/search**: Profile search and extraction
- **internal/stealth**: Anti-detection techniques
- **internal/storage**: SQLite persistence layer

## 🥷 Stealth Techniques

This project implements **11 comprehensive stealth techniques**:

### 1. Navigator.webdriver Masking
```javascript
Object.defineProperty(navigator, 'webdriver', {
    get: () => undefined
});
```
Removes the `navigator.webdriver` flag that identifies automated browsers.

### 2. Chrome Object Override
```javascript
window.chrome = { runtime: {} };
```
Adds the Chrome runtime object that's missing in headless mode.

### 3. Permissions API Override
Overrides the permissions query API to return realistic values.

### 4. Plugins Detection Override
Provides a realistic plugin array instead of empty.

### 5. Languages Override
Sets realistic language preferences.

### 6. Bezier Curve Mouse Movement ⭐
```go
// Cubic Bezier curve formula for natural mouse movement
x := math.Pow(1-t, 3)*startX +
    3*math.Pow(1-t, 2)*t*cp1X +
    3*(1-t)*math.Pow(t, 2)*cp2X +
    math.Pow(t, 3)*targetX
```
Moves the mouse along a Bezier curve with random control points, mimicking human movement.

### 7. Human Typing Simulation ⭐
- Random delays between keystrokes (100-300ms)
- 5% chance of typos with backspace correction
- Variable typing speed

### 8. Random Scrolling
- Scrolls up and down randomly
- Variable scroll distances (200-600px)
- Random delays between scrolls

### 9. Mouse Hovering and Wandering
- Random mouse movements to arbitrary positions
- Hover delays (500-2000ms)
- Simulates reading and exploration

### 10. Idle Breaks and Cool-down ⭐
- Automatic breaks every N actions (configurable)
- Break duration: 60-180 seconds
- Prevents sustained robotic activity

### 11. Reading Simulation
- Scrolls down page slowly as if reading
- Multiple scroll steps with delays
- 2-5 second pauses between scrolls

### Additional Stealth Features

- **Random Viewport Sizes**: Varies browser dimensions
- **Rotating User Agents**: Cycles through realistic user agents
- **Randomized Timing**: All delays are randomized within ranges
- **Business Hours Operation**: Only active during configured hours
- **Rate Limiting**: Enforces realistic daily/hourly limits

## ✨ Features

### Authentication
- ✅ Cookie-based session persistence
- ✅ Automatic session validation
- ✅ CAPTCHA detection with screenshot
- ✅ 2FA detection with manual intervention prompt
- ✅ Security challenge detection
- ✅ Login failure detection

### Search
- ✅ Multi-target search (job title, location, keywords)
- ✅ Pagination handling
- ✅ Profile data extraction (name, title, company)
- ✅ Deduplication
- ✅ Database persistence

### Connection Requests
- ✅ Personalized note templates
- ✅ Variable substitution ({{FirstName}}, {{Company}}, etc.)
- ✅ Note length validation
- ✅ Rate limiting (hourly/daily)
- ✅ Status tracking (pending/accepted/rejected)

### Messaging
- ✅ Automatic follow-up to accepted connections
- ✅ Configurable delay after connection
- ✅ Message templates with variables
- ✅ Message history tracking
- ✅ Rate limiting

### Scheduling
- ✅ Active hours configuration (e.g., 9 AM - 6 PM)
- ✅ Active days selection (weekdays only)
- ✅ Automatic waiting until next active period
- ✅ Business hours enforcement

### State Management
- ✅ SQLite database for all data
- ✅ Resume capability after interruption
- ✅ Activity logging for audit trail
- ✅ Statistics tracking (daily/hourly)

## 📁 Project Structure

```
linkedin-automation/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── auth/
│   │   └── auth.go            # Authentication service
│   ├── browser/
│   │   └── browser.go         # Browser context management
│   ├── config/
│   │   └── config.go          # Configuration loading
│   ├── connect/
│   │   └── connect.go         # Connection request service
│   ├── logger/
│   │   └── logger.go          # Logging setup
│   ├── message/
│   │   └── message.go         # Messaging service
│   ├── scheduler/
│   │   └── scheduler.go       # Activity scheduling
│   ├── search/
│   │   └── search.go          # Profile search service
│   ├── stealth/
│   │   └── stealth.go         # Stealth techniques
│   └── storage/
│       └── storage.go         # Database layer
├── config.yaml                 # Main configuration
├── .env.example               # Environment template
├── .gitignore                 # Git ignore rules
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── Makefile                   # Build automation
└── demo.sh                    # Demo walkthrough script
```

## 📦 Installation

### Prerequisites

1. **Go 1.21 or higher**
   ```bash
   # Check Go version
   go version
   ```

2. **Chromium/Chrome browser** (Rod will download if not found)

3. **Git** (for cloning)

### Steps

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd linkedin-automation
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your credentials
   ```

4. **Create required directories**
   ```bash
   mkdir -p data logs
   ```

5. **Build the application**
   ```bash
   go build -o linkedin-automation ./cmd/main.go
   ```

## ⚙️ Configuration

### Environment Variables (.env)

```env
# LinkedIn Credentials
LINKEDIN_EMAIL=your-email@example.com
LINKEDIN_PASSWORD=your-password-here

# Browser Settings
HEADLESS=false                    # Set to true for headless mode
CHROME_PATH=                      # Optional: custom Chrome path

# Rate Limiting
MAX_CONNECTIONS_PER_DAY=50
MAX_MESSAGES_PER_DAY=30
MAX_CONNECTIONS_PER_HOUR=10

# Stealth Settings
MIN_ACTION_DELAY_MS=2000
MAX_ACTION_DELAY_MS=5000

# Database
DB_PATH=./data/linkedin.db

# Logging
LOG_LEVEL=info                    # debug, info, warn, error
LOG_FILE=./logs/automation.log
```

### Configuration File (config.yaml)

See `config.yaml` for full configuration options including:

- Browser viewport and user agents
- Stealth technique toggles and timing
- Rate limits (connections, messages, searches)
- Search targets (job titles, locations, keywords)
- Connection note templates
- Message templates
- Active hours and days

## 🚀 Usage

### Basic Usage

```bash
# Run the application
./linkedin-automation
```

### Using Makefile

```bash
# Show available commands
make help

# Download dependencies
make deps

# Build the application
make build

# Run the application
make run

# Clean build artifacts
make clean

# Initialize project (first time)
make init
```

### Workflow

1. **Authentication**: Logs in to LinkedIn (or reuses session)
2. **Search**: Finds profiles matching configured targets
3. **Connect**: Sends connection requests with personalized notes
4. **Message**: Follows up with accepted connections
5. **Repeat**: Continues loop while respecting rate limits and schedule

### Graceful Shutdown

Press `Ctrl+C` to trigger graceful shutdown. The application will:
- Save current state
- Close browser cleanly
- Close database connections

## 🗄 Database Schema

### Tables

#### profiles
```sql
CREATE TABLE profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_url TEXT UNIQUE NOT NULL,
    name TEXT,
    job_title TEXT,
    company TEXT,
    location TEXT,
    keywords TEXT,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### connection_requests
```sql
CREATE TABLE connection_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id INTEGER,
    profile_url TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    note TEXT,
    status TEXT DEFAULT 'pending',
    accepted_at TIMESTAMP,
    FOREIGN KEY (profile_id) REFERENCES profiles(id)
);
```

#### messages
```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id INTEGER,
    profile_url TEXT NOT NULL,
    content TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'sent',
    FOREIGN KEY (profile_id) REFERENCES profiles(id)
);
```

#### activity_log
```sql
CREATE TABLE activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action_type TEXT NOT NULL,
    target_url TEXT,
    outcome TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 📊 Monitoring

### Logs

Logs are written to `./logs/automation.log` in structured format:

```
2024-01-15 10:30:45 [INFO] Starting LinkedIn Automation Bot
2024-01-15 10:30:46 [INFO] Authenticating with LinkedIn...
2024-01-15 10:30:52 [INFO] Authentication successful
2024-01-15 10:30:52 [INFO] Phase 1: Searching for target profiles...
```

### Database Queries

```sql
-- Today's statistics
SELECT 
    COUNT(*) FILTER (WHERE action_type = 'connection_request') as connections,
    COUNT(*) FILTER (WHERE action_type = 'message') as messages
FROM activity_log 
WHERE DATE(created_at) = DATE('now');

-- Recent connection requests
SELECT * FROM connection_requests 
ORDER BY sent_at DESC 
LIMIT 10;

-- Accepted connections not yet messaged
SELECT cr.* 
FROM connection_requests cr
LEFT JOIN messages m ON cr.profile_url = m.profile_url
WHERE cr.status = 'accepted' AND m.id IS NULL;
```

### Screenshots

On errors (CAPTCHA, 2FA, login failure), screenshots are saved to `./logs/`:
- `login_failed.png`
- `captcha_detected.png`
- `2fa_detected.png`
- `security_challenge.png`

## 🔧 Troubleshooting

### Common Issues

**1. "go: command not found"**
- Install Go from https://golang.org/dl/

**2. "failed to launch browser"**
- Ensure Chrome/Chromium is installed
- Set `CHROME_PATH` in .env if using custom location

**3. "CAPTCHA detected"**
- LinkedIn detected automation
- Check screenshot in `./logs/captcha_detected.png`
- May need to solve manually or wait before retrying

**4. "2FA verification required"**
- Your account has 2FA enabled
- Complete verification manually in the browser
- Session will be saved for future use

**5. "Rate limit reached"**
- Daily or hourly limit hit
- Wait for next period or adjust limits in config

**6. "Database locked"**
- Another instance is running
- Close other instances or delete `data/linkedin.db-journal`

### Debug Mode

Enable debug logging:
```env
LOG_LEVEL=debug
```

This will show detailed information about:
- Element selectors
- Mouse movements
- Timing delays
- Browser interactions

## 👨‍💻 Development

### Adding New Features

1. **New Service**: Create package in `internal/`
2. **Update Config**: Add configuration in `config/config.go`
3. **Update Main**: Integrate in `cmd/main.go`
4. **Test**: Add tests in `*_test.go` files

### Code Style

- Follow Go conventions
- Use structured logging
- Handle errors explicitly
- Add comments for complex logic
- Keep functions focused and small

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/stealth
```

## 📝 License

This project is provided for educational purposes only. See LICENSE file for details.

## 🙏 Acknowledgments

- [Rod](https://github.com/go-rod/rod) - Excellent browser automation library
- [Logrus](https://github.com/sirupsen/logrus) - Structured logging
- [SQLite](https://www.sqlite.org/) - Embedded database

---

**Remember**: This is a proof-of-concept for educational purposes. Always respect platform terms of service and user privacy.