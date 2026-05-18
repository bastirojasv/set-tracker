package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func openDB() (*DB, error) {
	dir, err := dbDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "settracker.db")
	conn, err := sql.Open("sqlite", path+"?_journal=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func dbDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA not set")
	}
	return filepath.Join(appData, "SetTracker"), nil
}

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
	CREATE TABLE IF NOT EXISTS user_profile (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL DEFAULT 'Jugador',
		color TEXT NOT NULL DEFAULT '#00BFFF',
		photo_path TEXT
	);

	CREATE TABLE IF NOT EXISTS sets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		played_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		set_type TEXT NOT NULL,
		set_number INTEGER NOT NULL,
		user_score INTEGER NOT NULL DEFAULT 0,
		opponent_score INTEGER NOT NULL DEFAULT 0,
		opponent_name TEXT NOT NULL,
		opponent_color TEXT NOT NULL DEFAULT '#FF4500',
		user_won INTEGER,
		status TEXT NOT NULL DEFAULT 'active'
	);

	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		set_id INTEGER NOT NULL REFERENCES sets(id),
		game_number INTEGER NOT NULL,
		winner TEXT NOT NULL
	);
	`)
	return err
}

// --- UserProfile ---

func (db *DB) HasProfile() (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM user_profile").Scan(&count)
	return count > 0, err
}

func (db *DB) LoadProfile() (*UserProfile, error) {
	row := db.conn.QueryRow("SELECT id, name, color, COALESCE(photo_path,'') FROM user_profile LIMIT 1")
	p := &UserProfile{}
	err := row.Scan(&p.ID, &p.Name, &p.Color, &p.PhotoPath)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (db *DB) SaveProfile(p *UserProfile) error {
	if p.ID == 0 {
		res, err := db.conn.Exec(
			"INSERT INTO user_profile(name, color, photo_path) VALUES(?,?,?)",
			p.Name, p.Color, nullStr(p.PhotoPath),
		)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		p.ID = int(id)
		return nil
	}
	_, err := db.conn.Exec(
		"UPDATE user_profile SET name=?, color=?, photo_path=? WHERE id=?",
		p.Name, p.Color, nullStr(p.PhotoPath), p.ID,
	)
	return err
}

// AbandonActiveSet marks any active set as abandoned (used for crash recovery on startup)
func (db *DB) AbandonActiveSet() {
	db.conn.Exec("UPDATE sets SET status='completed', user_won=0 WHERE status='active'")
}

// --- Active set ---

func (db *DB) LoadActiveSet() (*ActiveSetState, error) {
	row := db.conn.QueryRow(`
		SELECT id, set_type, set_number, user_score, opponent_score, opponent_name, opponent_color
		FROM sets WHERE status='active' ORDER BY id DESC LIMIT 1`)
	var s ActiveSetState
	err := row.Scan(&s.SetID, &s.SetType, &s.SetNumber,
		&s.UserScore, &s.OppScore, &s.OpponentName, &s.OpponentColor)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Load games
	rows, err := db.conn.Query(
		"SELECT winner FROM games WHERE set_id=? ORDER BY game_number", s.SetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var w string
		rows.Scan(&w)
		s.Games = append(s.Games, w)
	}
	return &s, nil
}

func (db *DB) StartSet(setType string, setNum int, oppName, oppColor string) (*ActiveSetState, error) {
	// Cancel any previous active set (shouldn't happen, but be safe)
	db.conn.Exec("UPDATE sets SET status='completed', user_won=0 WHERE status='active'")

	res, err := db.conn.Exec(
		"INSERT INTO sets(set_type, set_number, opponent_name, opponent_color, played_at) VALUES(?,?,?,?,?)",
		setType, setNum, oppName, oppColor, time.Now().Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &ActiveSetState{
		SetID:         int(id),
		SetType:       setType,
		SetNumber:     setNum,
		OpponentName:  oppName,
		OpponentColor: oppColor,
	}, nil
}

func (db *DB) RecordGame(s *ActiveSetState, winner string) error {
	gameNum := len(s.Games) + 1
	_, err := db.conn.Exec(
		"INSERT INTO games(set_id, game_number, winner) VALUES(?,?,?)",
		s.SetID, gameNum, winner,
	)
	if err != nil {
		return err
	}
	_, err = db.conn.Exec(
		"UPDATE sets SET user_score=?, opponent_score=? WHERE id=?",
		s.UserScore, s.OppScore, s.SetID,
	)
	return err
}

func (db *DB) UndoLastGame(s *ActiveSetState) error {
	if len(s.Games) == 0 {
		return nil
	}
	gameNum := len(s.Games)
	_, err := db.conn.Exec(
		"DELETE FROM games WHERE set_id=? AND game_number=?",
		s.SetID, gameNum,
	)
	if err != nil {
		return err
	}
	_, err = db.conn.Exec(
		"UPDATE sets SET user_score=?, opponent_score=? WHERE id=?",
		s.UserScore, s.OppScore, s.SetID,
	)
	return err
}

func (db *DB) CompleteSet(s *ActiveSetState, userWon bool) error {
	won := 0
	if userWon {
		won = 1
	}
	_, err := db.conn.Exec(
		"UPDATE sets SET status='completed', user_won=?, user_score=?, opponent_score=? WHERE id=?",
		won, s.UserScore, s.OppScore, s.SetID,
	)
	return err
}

func (db *DB) UpdateSetType(s *ActiveSetState, setType string, setNum int) error {
	_, err := db.conn.Exec(
		"UPDATE sets SET set_type=?, set_number=? WHERE id=?",
		setType, setNum, s.SetID,
	)
	return err
}

// --- History ---

func (db *DB) LoadHistory() ([]SetRecord, error) {
	rows, err := db.conn.Query(`
		SELECT id, played_at, set_type, set_number,
			user_score, opponent_score, opponent_name, opponent_color, user_won
		FROM sets WHERE status='completed'
		ORDER BY played_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []SetRecord
	for rows.Next() {
		var r SetRecord
		var wonInt *int
		var playedAt string
		err := rows.Scan(&r.ID, &playedAt, &r.SetType, &r.SetNumber,
			&r.UserScore, &r.OpponentScore, &r.OpponentName, &r.OpponentColor, &wonInt)
		if err != nil {
			continue
		}
		r.PlayedAt = parseDateTime(playedAt)
		if wonInt != nil {
			b := *wonInt == 1
			r.UserWon = &b
		}
		records = append(records, r)
	}
	return records, nil
}

func (db *DB) LoadOpponents() ([]string, error) {
	rows, err := db.conn.Query(`
		SELECT DISTINCT opponent_name FROM sets
		ORDER BY opponent_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	return names, nil
}

func (db *DB) LoadOpponentColor(name string) string {
	var c string
	db.conn.QueryRow(
		"SELECT opponent_color FROM sets WHERE opponent_name=? ORDER BY id DESC LIMIT 1", name,
	).Scan(&c)
	if c == "" {
		c = "#FF4500"
	}
	return c
}

func (db *DB) LoadH2H(opponent string) ([]SetRecord, error) {
	rows, err := db.conn.Query(`
		SELECT id, played_at, set_type, set_number,
			user_score, opponent_score, opponent_name, opponent_color, user_won
		FROM sets WHERE status='completed' AND opponent_name=?
		ORDER BY played_at DESC`, opponent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []SetRecord
	for rows.Next() {
		var r SetRecord
		var wonInt *int
		var playedAt string
		rows.Scan(&r.ID, &playedAt, &r.SetType, &r.SetNumber,
			&r.UserScore, &r.OpponentScore, &r.OpponentName, &r.OpponentColor, &wonInt)
		r.PlayedAt = parseDateTime(playedAt)
		if wonInt != nil {
			b := *wonInt == 1
			r.UserWon = &b
		}
		records = append(records, r)
	}
	return records, nil
}

// --- Deletion ---

func (db *DB) DeleteSet(id int) error {
	if _, err := db.conn.Exec("DELETE FROM games WHERE set_id=?", id); err != nil {
		return err
	}
	_, err := db.conn.Exec("DELETE FROM sets WHERE id=?", id)
	return err
}

func (db *DB) DeleteHistoryByOpponent(name string) error {
	if _, err := db.conn.Exec(
		"DELETE FROM games WHERE set_id IN (SELECT id FROM sets WHERE status='completed' AND opponent_name=?)", name,
	); err != nil {
		return err
	}
	_, err := db.conn.Exec("DELETE FROM sets WHERE status='completed' AND opponent_name=?", name)
	return err
}

func (db *DB) DeleteAllHistory() error {
	if _, err := db.conn.Exec(
		"DELETE FROM games WHERE set_id IN (SELECT id FROM sets WHERE status='completed')",
	); err != nil {
		return err
	}
	_, err := db.conn.Exec("DELETE FROM sets WHERE status='completed'")
	return err
}

// parseDateTime handles the multiple formats that modernc/sqlite may produce
// when storing time.Time values (RFC3339Nano, RFC3339, plain datetime string).
func parseDateTime(s string) time.Time {
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
