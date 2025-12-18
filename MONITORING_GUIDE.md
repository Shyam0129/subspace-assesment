# How to Check Logs and Monitor LinkedIn Automation

## 📊 Viewing Logs

### Real-Time Log Monitoring

Watch logs as they're written (like `tail -f` on Linux):

```powershell
Get-Content .\logs\automation.log -Tail 50 -Wait
```

This will show the last 50 lines and keep updating as new logs are written.

### View All Logs

```powershell
Get-Content .\logs\automation.log
```

### View Last N Lines

```powershell
# Last 20 lines
Get-Content .\logs\automation.log -Tail 20

# Last 100 lines
Get-Content .\logs\automation.log -Tail 100
```

### Search Logs for Errors

```powershell
# Find all error messages
Get-Content .\logs\automation.log | Select-String "error" -CaseSensitive:$false

# Find all fatal messages
Get-Content .\logs\automation.log | Select-String "fatal" -CaseSensitive:$false

# Find specific action
Get-Content .\logs\automation.log | Select-String "connection_request"
```

### Filter Logs by Time

```powershell
# Get logs from today
Get-Content .\logs\automation.log | Select-String (Get-Date -Format "yyyy-MM-dd")
```

## 📁 Log File Location

- **Path**: `.\logs\automation.log`
- **Format**: Text file with timestamps
- **Rotation**: Currently no rotation (file grows indefinitely)

## 🗄️ Database Monitoring

### Check if Database Exists

```powershell
Test-Path .\data\linkedin.db
```

### View Database File Info

```powershell
Get-Item .\data\linkedin.db | Select-Object Name, Length, LastWriteTime
```

### Query Database (Requires SQLite)

If you have SQLite installed:

```powershell
# Open database
sqlite3 .\data\linkedin.db

# Inside SQLite prompt:
# .tables                          # List all tables
# SELECT * FROM profiles LIMIT 10; # View profiles
# SELECT * FROM activity_log ORDER BY created_at DESC LIMIT 20; # Recent activity
# .quit                            # Exit
```

### Install SQLite (if needed)

```powershell
# Using winget
winget install SQLite.SQLite

# Or download from: https://www.sqlite.org/download.html
```

## 📸 Screenshot Monitoring

The application saves screenshots when errors occur:

```powershell
# List all screenshots
Get-ChildItem .\logs\*.png | Select-Object Name, Length, LastWriteTime

# View specific screenshots
# captcha_detected.png - CAPTCHA challenge detected
# 2fa_detected.png - 2FA verification required
# login_failed.png - Login failure
# security_challenge.png - Security challenge detected
```

Open screenshots:

```powershell
# Open in default image viewer
Start-Process .\logs\captcha_detected.png
```

## 🔍 Common Log Messages

### Successful Startup

```
time="2025-12-18 17:50:00" level=info msg="Starting LinkedIn Automation Bot"
time="2025-12-18 17:50:01" level=info msg="Initializing browser..."
time="2025-12-18 17:50:05" level=info msg="Browser initialized successfully"
time="2025-12-18 17:50:05" level=info msg="Authenticating with LinkedIn..."
```

### Authentication Success

```
time="2025-12-18 17:50:10" level=info msg="Authentication successful"
time="2025-12-18 17:50:10" level=info msg="Starting profile search..."
```

### Connection Request Sent

```
time="2025-12-18 17:51:00" level=info msg="Sending connection request to: https://linkedin.com/in/username"
time="2025-12-18 17:51:05" level=info msg="Connection request sent to John Doe (1/10)"
```

### Rate Limit Reached

```
time="2025-12-18 18:00:00" level=warn msg="Hourly connection limit reached"
time="2025-12-18 18:00:00" level=info msg="Sent 10 connection requests"
```

### Errors to Watch For

```
level=error msg="CAPTCHA detected"
level=error msg="2FA verification required"
level=error msg="Failed to navigate to profile"
level=fatal msg="Failed to initialize storage"
```

## 📈 Monitoring Dashboard (PowerShell)

Create a simple monitoring script:

```powershell
# Save as monitor.ps1
while ($true) {
    Clear-Host
    Write-Host "=== LinkedIn Automation Monitor ===" -ForegroundColor Cyan
    Write-Host ""
    
    # Check if app is running
    $process = Get-Process -Name "linkedin-automation" -ErrorAction SilentlyContinue
    if ($process) {
        Write-Host "Status: RUNNING (PID: $($process.Id))" -ForegroundColor Green
    } else {
        Write-Host "Status: NOT RUNNING" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "=== Recent Logs ===" -ForegroundColor Yellow
    Get-Content .\logs\automation.log -Tail 10 -ErrorAction SilentlyContinue
    
    Write-Host ""
    Write-Host "=== Database Info ===" -ForegroundColor Yellow
    if (Test-Path .\data\linkedin.db) {
        $db = Get-Item .\data\linkedin.db
        Write-Host "Size: $([math]::Round($db.Length/1KB, 2)) KB"
        Write-Host "Last Modified: $($db.LastWriteTime)"
    } else {
        Write-Host "Database not found" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "Press Ctrl+C to exit. Refreshing in 5 seconds..." -ForegroundColor Gray
    Start-Sleep -Seconds 5
}
```

Run it:

```powershell
.\monitor.ps1
```

## 🎯 Quick Checks

### Is the app running?

```powershell
Get-Process -Name "linkedin-automation" -ErrorAction SilentlyContinue
```

### Last 10 log entries

```powershell
Get-Content .\logs\automation.log -Tail 10
```

### Check for errors in last run

```powershell
Get-Content .\logs\automation.log | Select-String "error|fatal" -CaseSensitive:$false | Select-Object -Last 10
```

### Count today's activities

```powershell
$today = Get-Date -Format "yyyy-MM-dd"
$logs = Get-Content .\logs\automation.log
$connections = ($logs | Select-String "Connection request sent").Count
$messages = ($logs | Select-String "Message sent").Count

Write-Host "Today's Stats:"
Write-Host "Connections: $connections"
Write-Host "Messages: $messages"
```

## 🔧 Troubleshooting

### No logs appearing?

Check if the logs directory exists:

```powershell
Test-Path .\logs
# If false, create it:
New-Item -ItemType Directory -Path .\logs
```

### Log file too large?

Archive old logs:

```powershell
# Rename current log
Move-Item .\logs\automation.log .\logs\automation_$(Get-Date -Format 'yyyyMMdd_HHmmss').log

# App will create new log file on next run
```

### Can't read database?

Check file permissions:

```powershell
Get-Acl .\data\linkedin.db | Format-List
```

## 📊 Log Levels

The application uses these log levels:

- **DEBUG**: Detailed information for debugging (only if LOG_LEVEL=debug in .env)
- **INFO**: General informational messages (default)
- **WARN**: Warning messages (rate limits, retries, etc.)
- **ERROR**: Error messages (recoverable errors)
- **FATAL**: Fatal errors (application will exit)

### Change Log Level

Edit `.env`:

```env
# For more detailed logs
LOG_LEVEL=debug

# For normal operation
LOG_LEVEL=info

# For warnings and errors only
LOG_LEVEL=warn
```

## 🎉 Success Indicators

You'll know the app is working correctly when you see:

1. ✅ "Starting LinkedIn Automation Bot"
2. ✅ "Browser initialized successfully"
3. ✅ "Authentication successful"
4. ✅ "Starting profile search..."
5. ✅ "Connection request sent to..."
6. ✅ Database file created in `.\data\`
7. ✅ No FATAL errors in logs

## ⚠️ Warning Signs

Watch out for these in logs:

- ❌ "CAPTCHA detected" - LinkedIn detected automation
- ❌ "2FA verification required" - Manual intervention needed
- ❌ "Rate limit reached" - Too many actions (expected behavior)
- ❌ "Failed to initialize storage" - Database issue
- ❌ "Failed to launch browser" - Browser/Chrome issue

---

**Pro Tip**: Keep a terminal window open with `Get-Content .\logs\automation.log -Tail 50 -Wait` to monitor in real-time!
