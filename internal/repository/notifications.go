package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/internal/notifications"
)

type Notification struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Severity  int       `json:"severity"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	RelatedID string    `json:"related_id"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
}

func readOnlySQLiteURI(path string) string {
	path = filepath.ToSlash(path)
	uri := &url.URL{Scheme: "file"}
	if runtime.GOOS == "windows" && strings.HasPrefix(path, "//") {
		unc := strings.TrimPrefix(path, "//")
		uri.Host, path, _ = strings.Cut(unc, "/")
		path = "/" + path
	} else if runtime.GOOS == "windows" && len(path) > 1 && path[1] == ':' {
		path = "/" + path
	}
	uri.Path = path
	query := uri.Query()
	query.Set("mode", "ro")
	uri.RawQuery = query.Encode()
	return uri.String()
}

func OpenReadOnly(path string) (*Store, error) {
	uri := readOnlySQLiteURI(path)
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, fmt.Errorf("open read-only database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set read timeout: %w", err)
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("read-only database is unavailable: %w", err)
	}
	return &Store{db: db, path: path}, nil
}

func (s *Store) AddNotification(ctx context.Context, event notifications.Event, window time.Duration) (bool, error) {
	key := notifications.Key(event)
	now := time.Now().UTC()
	var last string
	err := s.db.QueryRowContext(ctx, "SELECT created_at FROM notifications WHERE dedup_key=? ORDER BY id DESC LIMIT 1", key).Scan(&last)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("check notification duplicates: %w", err)
	}
	if err == nil {
		parsed, parseErr := time.Parse(time.RFC3339Nano, last)
		if parseErr == nil && notifications.Deduplicated(parsed, now, window) {
			return false, nil
		}
	}
	_, err = s.db.ExecContext(ctx, "INSERT INTO notifications(event_type,severity,title,detail,related_id,dedup_key,created_at) VALUES(?,?,?,?,?,?,?)", event.Type, event.Severity, event.Title, event.Detail, event.Related, key, now.Format(time.RFC3339Nano))
	if err != nil {
		return false, fmt.Errorf("save notification: %w", err)
	}
	return true, nil
}

func (s *Store) Notifications(ctx context.Context, limit int) ([]Notification, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, "SELECT id,event_type,severity,title,detail,related_id,created_at,read_at FROM notifications ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("read notifications: %w", err)
	}
	defer rows.Close()
	items := make([]Notification, 0)
	for rows.Next() {
		var item Notification
		var created, readAt string
		if err = rows.Scan(&item.ID, &item.Type, &item.Severity, &item.Title, &item.Detail, &item.RelatedID, &created, &readAt); err != nil {
			return nil, fmt.Errorf("read notification: %w", err)
		}
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, fmt.Errorf("parse notification time: %w", err)
		}
		item.Read = readAt != ""
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UnreadCount(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notifications WHERE read_at='' ").Scan(&count); err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}

func (s *Store) MarkNotificationRead(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "UPDATE notifications SET read_at=? WHERE id=?", time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

func (s *Store) MarkAllNotificationsRead(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "UPDATE notifications SET read_at=? WHERE read_at=''", time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("mark notifications read: %w", err)
	}
	return nil
}

func (s *Store) NotificationRules(ctx context.Context) (map[string]notifications.Rule, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT event_type,enabled,minimum_severity,quiet_start,quiet_end,detailed_text FROM notification_rules ORDER BY event_type")
	if err != nil {
		return nil, fmt.Errorf("read notification rules: %w", err)
	}
	defer rows.Close()
	rules := make(map[string]notifications.Rule)
	for rows.Next() {
		var eventType string
		var enabled, severity, detailed int
		var start, end string
		if err = rows.Scan(&eventType, &enabled, &severity, &start, &end, &detailed); err != nil {
			return nil, err
		}
		rules[eventType] = notifications.Rule{Enabled: enabled != 0, MinimumSeverity: severity, QuietStart: start, QuietEnd: end, DetailedText: detailed != 0}
	}
	return rules, rows.Err()
}

func (s *Store) SaveNotificationRule(ctx context.Context, eventType string, rule notifications.Rule) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO notification_rules(event_type,enabled,minimum_severity,quiet_start,quiet_end,detailed_text) VALUES(?,?,?,?,?,?) ON CONFLICT(event_type) DO UPDATE SET enabled=excluded.enabled,minimum_severity=excluded.minimum_severity,quiet_start=excluded.quiet_start,quiet_end=excluded.quiet_end,detailed_text=excluded.detailed_text", eventType, boolInt(rule.Enabled), rule.MinimumSeverity, rule.QuietStart, rule.QuietEnd, boolInt(rule.DetailedText))
	if err != nil {
		return fmt.Errorf("save notification rule: %w", err)
	}
	return nil
}

func (s *Store) PendingNotifications(ctx context.Context) ([]Notification, error) {
	rules, err := s.NotificationRules(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, "SELECT id,event_type,severity,title,detail,related_id,created_at,read_at FROM notifications WHERE delivered_at='' AND read_at='' ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("read pending notifications: %w", err)
	}
	defer rows.Close()
	now := time.Now()
	items := make([]Notification, 0)
	for rows.Next() {
		var item Notification
		var created, readAt string
		if err = rows.Scan(&item.ID, &item.Type, &item.Severity, &item.Title, &item.Detail, &item.RelatedID, &created, &readAt); err != nil {
			return nil, fmt.Errorf("read pending notification: %w", err)
		}
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, fmt.Errorf("parse notification time: %w", err)
		}
		item.Read = readAt != ""
		rule, ok := rules[item.Type]
		if ok && notifications.Allowed(notifications.Event{Severity: item.Severity}, rule, now) {
			items = append(items, item)
		}
	}
	return items, rows.Err()
}

func (s *Store) MarkNotificationDelivered(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "UPDATE notifications SET delivered_at=? WHERE id=?", time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("mark notification delivered: %w", err)
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
