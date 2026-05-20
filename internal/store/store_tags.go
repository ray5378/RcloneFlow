package store

func (db *DB) ListTags() ([]Tag, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query(`SELECT id, tag, type FROM task_tags ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Tag, &t.Type); err != nil {
			continue
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (db *DB) ReplaceTags(tags map[string]string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := db.db.Exec(`DELETE FROM task_tags`); err != nil {
		return err
	}

	for tag, tagType := range tags {
		if _, err := db.db.Exec(`INSERT INTO task_tags (tag, type) VALUES (?, ?)`, tag, tagType); err != nil {
			return err
		}
	}
	return nil
}