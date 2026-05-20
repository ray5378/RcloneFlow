package store

import "time"

func (db *DB) CreateUser(username, password string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.db.Exec(`
		INSERT INTO users (username, password, password_changed) VALUES (?, ?, 0)`,
		username, password)
	if err != nil {
		return User{}, err
	}

	id, _ := result.LastInsertId()
	return User{
		ID:              id,
		Username:        username,
		Password:        password,
		PasswordChanged: false,
		CreatedAt:       time.Now(),
	}, nil
}

func (db *DB) GetUserByUsername(username string) (User, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var u User
	err := db.db.QueryRow(`
		SELECT id, username, password, password_changed, created_at FROM users WHERE username = ?`, username).
		Scan(&u.ID, &u.Username, &u.Password, &u.PasswordChanged, &u.CreatedAt)
	if err != nil {
		return User{}, false
	}
	return u, true
}

func (db *DB) GetUserByID(id int64) (User, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var u User
	err := db.db.QueryRow(`
		SELECT id, username, password, password_changed, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Username, &u.Password, &u.PasswordChanged, &u.CreatedAt)
	if err != nil {
		return User{}, false
	}
	return u, true
}

func (db *DB) UpdatePassword(id int64, hashedPassword string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`UPDATE users SET password = ?, password_changed = 1 WHERE id = ?`, hashedPassword, id)
	return err
}

func (db *DB) UpdateUsername(id int64, username string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`UPDATE users SET username = ?, password_changed = 1 WHERE id = ?`, username, id)
	return err
}

func (db *DB) ListUsers() ([]User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query(`SELECT id, username, password, password_changed, created_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.PasswordChanged, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}