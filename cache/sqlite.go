package cache

import (
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Cache struct {
	db         *sql.DB
	maxEntries int
	mu         sync.Mutex
}

func New(dbPath string, maxEntries int) (*Cache, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.Exec("PRAGMA journal_mode=WAL")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS routes (
		key TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		created_at INTEGER NOT NULL
	)`)
	if err != nil {
		return nil, err
	}

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_routes_created ON routes(created_at)`)

	return &Cache{db: db, maxEntries: maxEntries}, nil
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var data []byte
	err := c.db.QueryRow("SELECT data FROM routes WHERE key = ?", key).Scan(&data)
	if err != nil {
		return nil, false
	}

	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}
	return result, true
}

func (c *Cache) Put(key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var count int
	c.db.QueryRow("SELECT COUNT(*) FROM routes").Scan(&count)
	if count >= c.maxEntries {
		excess := count - c.maxEntries + 1
		c.db.Exec("DELETE FROM routes WHERE key IN (SELECT key FROM routes ORDER BY created_at ASC LIMIT ?)", excess)
	}

	_, err = c.db.Exec("INSERT OR REPLACE INTO routes (key, data, created_at) VALUES (?, ?, ?)",
		key, data, time.Now().Unix())
	return err
}

func (c *Cache) Close() error {
	return c.db.Close()
}
