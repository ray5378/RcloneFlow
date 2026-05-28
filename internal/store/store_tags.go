package store

import "strings"

type TagEntry struct {
	Tag      string
	Type     string
	Selected bool
}

func (db *DB) ListTags() ([]Tag, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query(`SELECT id, tag, type, selected FROM task_tags ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Tag, &t.Type, &t.Selected); err != nil {
			continue
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (db *DB) MergeTags(entries []TagEntry) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := db.db.Exec(`DELETE FROM task_tags WHERE type = 'keyword'`); err != nil {
		return err
	}

	stmt, err := db.db.Prepare(`INSERT INTO task_tags (tag, type, selected) VALUES (?, ?, ?) ON CONFLICT(tag) DO UPDATE SET type = excluded.type`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range entries {
		selected := 0
		if e.Selected {
			selected = 1
		}
		if _, err := stmt.Exec(e.Tag, e.Type, selected); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) SetTagSelected(tag string, selected bool) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	val := 0
	if selected {
		val = 1
	}
	_, err := db.db.Exec(`UPDATE task_tags SET selected = ? WHERE tag = ?`, val, tag)
	return err
}

func (db *DB) CreateManualTag(tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`INSERT INTO task_tags (tag, type, selected) VALUES (?, 'keyword', 1) ON CONFLICT(tag) DO UPDATE SET selected = 1`, tag)
	return err
}

func (db *DB) DeleteManualTag(tag string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`DELETE FROM task_tags WHERE tag = ?`, tag)
	return err
}