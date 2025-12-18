package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

type Profile struct {
	ID           int64
	ProfileURL   string
	Name         string
	JobTitle     string
	Company      string
	Location     string
	Keywords     string
	DiscoveredAt time.Time
}

type ConnectionRequest struct {
	ID         int64
	ProfileID  int64
	ProfileURL string
	SentAt     time.Time
	Note       string
	Status     string // pending, accepted, rejected
	AcceptedAt *time.Time
}

type Message struct {
	ID         int64
	ProfileID  int64
	ProfileURL string
	Content    string
	SentAt     time.Time
	Status     string // sent, failed
}

type DailyStats struct {
	ConnectionsSent   int
	MessagesSent      int
	SearchesPerformed int
}

// New creates a new storage instance
func New(dbPath string) (*Storage, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the database schema
func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_url TEXT UNIQUE NOT NULL,
		name TEXT,
		job_title TEXT,
		company TEXT,
		location TEXT,
		keywords TEXT,
		discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS connection_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_id INTEGER,
		profile_url TEXT NOT NULL,
		sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		note TEXT,
		status TEXT DEFAULT 'pending',
		accepted_at TIMESTAMP,
		FOREIGN KEY (profile_id) REFERENCES profiles(id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_id INTEGER,
		profile_url TEXT NOT NULL,
		content TEXT NOT NULL,
		sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'sent',
		FOREIGN KEY (profile_id) REFERENCES profiles(id)
	);

	CREATE TABLE IF NOT EXISTS activity_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action_type TEXT NOT NULL,
		target_url TEXT,
		outcome TEXT,
		error_message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_profiles_url ON profiles(profile_url);
	CREATE INDEX IF NOT EXISTS idx_connections_status ON connection_requests(status);
	CREATE INDEX IF NOT EXISTS idx_connections_sent_at ON connection_requests(sent_at);
	CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages(sent_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveProfile saves a profile to the database
func (s *Storage) SaveProfile(profile *Profile) (int64, error) {
	result, err := s.db.Exec(`
		INSERT OR IGNORE INTO profiles (profile_url, name, job_title, company, location, keywords)
		VALUES (?, ?, ?, ?, ?, ?)
	`, profile.ProfileURL, profile.Name, profile.JobTitle, profile.Company, profile.Location, profile.Keywords)

	if err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()
	if id == 0 {
		// Profile already exists, get its ID
		err = s.db.QueryRow("SELECT id FROM profiles WHERE profile_url = ?", profile.ProfileURL).Scan(&id)
	}

	return id, err
}

// SaveConnectionRequest saves a connection request
func (s *Storage) SaveConnectionRequest(req *ConnectionRequest) error {
	_, err := s.db.Exec(`
		INSERT INTO connection_requests (profile_id, profile_url, note, status)
		VALUES (?, ?, ?, ?)
	`, req.ProfileID, req.ProfileURL, req.Note, req.Status)

	return err
}

// SaveMessage saves a message
func (s *Storage) SaveMessage(msg *Message) error {
	_, err := s.db.Exec(`
		INSERT INTO messages (profile_id, profile_url, content, status)
		VALUES (?, ?, ?, ?)
	`, msg.ProfileID, msg.ProfileURL, msg.Content, msg.Status)

	return err
}

// IsConnectionSent checks if a connection request was already sent to a profile
func (s *Storage) IsConnectionSent(profileURL string) (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM connection_requests WHERE profile_url = ?
	`, profileURL).Scan(&count)

	return count > 0, err
}

// IsMessageSent checks if a message was already sent to a profile
func (s *Storage) IsMessageSent(profileURL string) (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM messages WHERE profile_url = ?
	`, profileURL).Scan(&count)

	return count > 0, err
}

// GetAcceptedConnections returns connections that were accepted and haven't been messaged
func (s *Storage) GetAcceptedConnections() ([]ConnectionRequest, error) {
	rows, err := s.db.Query(`
		SELECT cr.id, cr.profile_id, cr.profile_url, cr.sent_at, cr.note, cr.status, cr.accepted_at
		FROM connection_requests cr
		LEFT JOIN messages m ON cr.profile_url = m.profile_url
		WHERE cr.status = 'accepted' AND m.id IS NULL
		ORDER BY cr.accepted_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []ConnectionRequest
	for rows.Next() {
		var conn ConnectionRequest
		if err := rows.Scan(&conn.ID, &conn.ProfileID, &conn.ProfileURL, &conn.SentAt, &conn.Note, &conn.Status, &conn.AcceptedAt); err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// GetTodayStats returns statistics for today
func (s *Storage) GetTodayStats() DailyStats {
	var stats DailyStats
	today := time.Now().Format("2006-01-02")

	s.db.QueryRow(`
		SELECT COUNT(*) FROM connection_requests 
		WHERE DATE(sent_at) = ?
	`, today).Scan(&stats.ConnectionsSent)

	s.db.QueryRow(`
		SELECT COUNT(*) FROM messages 
		WHERE DATE(sent_at) = ?
	`, today).Scan(&stats.MessagesSent)

	return stats
}

// GetHourlyStats returns statistics for the current hour
func (s *Storage) GetHourlyStats() DailyStats {
	var stats DailyStats
	hourAgo := time.Now().Add(-1 * time.Hour)

	s.db.QueryRow(`
		SELECT COUNT(*) FROM connection_requests 
		WHERE sent_at >= ?
	`, hourAgo).Scan(&stats.ConnectionsSent)

	s.db.QueryRow(`
		SELECT COUNT(*) FROM messages 
		WHERE sent_at >= ?
	`, hourAgo).Scan(&stats.MessagesSent)

	return stats
}

// LogActivity logs an activity to the database
func (s *Storage) LogActivity(actionType, targetURL, outcome, errorMessage string) error {
	_, err := s.db.Exec(`
		INSERT INTO activity_log (action_type, target_url, outcome, error_message)
		VALUES (?, ?, ?, ?)
	`, actionType, targetURL, outcome, errorMessage)

	return err
}

// UpdateConnectionStatus updates the status of a connection request
func (s *Storage) UpdateConnectionStatus(profileURL, status string) error {
	_, err := s.db.Exec(`
		UPDATE connection_requests 
		SET status = ?, accepted_at = CASE WHEN ? = 'accepted' THEN CURRENT_TIMESTAMP ELSE accepted_at END
		WHERE profile_url = ?
	`, status, status, profileURL)

	return err
}

// Close closes the database connection
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetProfileByURL retrieves a profile by URL
func (s *Storage) GetProfileByURL(url string) (*Profile, error) {
	var profile Profile
	err := s.db.QueryRow(`
		SELECT id, profile_url, name, job_title, company, location, keywords, discovered_at
		FROM profiles WHERE profile_url = ?
	`, url).Scan(&profile.ID, &profile.ProfileURL, &profile.Name, &profile.JobTitle,
		&profile.Company, &profile.Location, &profile.Keywords, &profile.DiscoveredAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &profile, err
}
