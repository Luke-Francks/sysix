package metricstore

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Point struct {
	Time        int64
	CPUPercent  float64
	MemPercent  float64
	DiskPercent float64
}

// Store persists rolling metrics samples per host in SQLite.
type Store struct {
	db        *sql.DB
	retention time.Duration
	apiLimit  int

	mu        sync.Mutex
	lastPrune time.Time
}

func Open(dbPath string, retentionDays int, apiLimit int) (*Store, error) {
	if retentionDays < 1 {
		retentionDays = 7
	}
	if apiLimit < 2 {
		apiLimit = 60
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS metrics (
  host_id TEXT NOT NULL,
  ts INTEGER NOT NULL,
  cpu REAL NOT NULL,
  mem REAL NOT NULL,
  disk REAL NOT NULL,
  PRIMARY KEY (host_id, ts)
);
CREATE INDEX IF NOT EXISTS metrics_host_ts ON metrics(host_id, ts);
`); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{
		db:        db,
		retention: time.Duration(retentionDays) * 24 * time.Hour,
		apiLimit:  apiLimit,
	}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) APILimit() int {
	if s == nil {
		return 60
	}
	return s.apiLimit
}

// Insert records a sample. Best-effort prune runs at most once per minute.
func (s *Store) Insert(hostID string, p Point) error {
	if s == nil || s.db == nil {
		return nil
	}
	if hostID == "" {
		hostID = "local"
	}
	_, err := s.db.Exec(`INSERT OR REPLACE INTO metrics (host_id, ts, cpu, mem, disk) VALUES (?, ?, ?, ?, ?)`,
		hostID, p.Time, p.CPUPercent, p.MemPercent, p.DiskPercent)
	if err != nil {
		return err
	}
	s.maybePrune()
	return nil
}

func (s *Store) maybePrune() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Since(s.lastPrune) < time.Minute {
		return
	}
	cutoff := time.Now().Add(-s.retention).Unix()
	if _, err := s.db.Exec(`DELETE FROM metrics WHERE ts < ?`, cutoff); err == nil {
		s.lastPrune = time.Now()
	}
}

// Recent returns up to limit points in chronological order (oldest first).
func (s *Store) Recent(hostID string, limit int) ([]Point, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	if hostID == "" {
		hostID = "local"
	}
	if limit < 2 {
		limit = s.apiLimit
	}

	rows, err := s.db.Query(`
SELECT ts, cpu, mem, disk FROM metrics
 WHERE host_id = ?
 ORDER BY ts DESC
 LIMIT ?`, hostID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pts []Point
	for rows.Next() {
		var p Point
		if err := rows.Scan(&p.Time, &p.CPUPercent, &p.MemPercent, &p.DiskPercent); err != nil {
			return nil, err
		}
		pts = append(pts, p)
	}
	// reverse to chronological
	for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
		pts[i], pts[j] = pts[j], pts[i]
	}
	return pts, nil
}
