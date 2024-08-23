package RebootForums

import (
	"database/sql"
	"log"
_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return err
	}

	err = DB.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return err
	}

	log.Println("Database connection established")
	return nil
}

// CreateTables creates all the necessary tables if they don't exist
func CreateTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER,
			user_id INTEGER,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS post_categories (
			post_id INTEGER,
			category_id INTEGER,
			PRIMARY KEY (post_id, category_id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,
		`CREATE TABLE IF NOT EXISTS likes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			post_id INTEGER,
			comment_id INTEGER,
			is_like BOOLEAN NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (comment_id) REFERENCES comments(id),
			UNIQUE(user_id, post_id, comment_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			token TEXT UNIQUE NOT NULL,
			expiry DATETIME NOT NULL,
			is_guest BOOLEAN NOT NULL DEFAULT 0,
			last_activity DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Printf("Error executing query: %s\nError: %v", query, err)
			return err
		}
	}

	log.Println("Tables created successfully")

	// Add default categories
	err := addDefaultCategories()
	if err != nil {
		log.Printf("Error adding default categories: %v", err)
		return err
	}

	return nil
}

// addDefaultCategories adds default categories to the database
func addDefaultCategories() error {
	categories := []string{
		"General Discussion",
		"Technology",
		"Sports",
		"Entertainment",
		"Science",
		"Politics",
		"Health",
		"Education",
		"Travel",
		"Food",
	}

	for _, category := range categories {
		_, err := DB.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", category)
		if err != nil {
			return err
		}
	}

	log.Println("Default categories added successfully")
	return nil
}

// AddUpdatedAtColumn adds the updated_at column to the posts table if it doesn't exist
func AddUpdatedAtColumn() error {
	_, err := DB.Exec(`
		ALTER TABLE posts ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	`)
	if err != nil {
		// If the error is because the column already exists, we can ignore it
		if err.Error() != "duplicate column name: updated_at" {
			log.Printf("Error adding updated_at column: %v", err)
			return err
		}
	}
	log.Println("updated_at column added to posts table (if it didn't exist)")
	return nil
}

// GetLikeCounts returns the number of likes and dislikes for a post or comment
func GetLikeCounts(targetID int, isPost bool) (likes int, dislikes int, err error) {
    var query string
    if isPost {
        query = `
            SELECT 
                COALESCE(SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END), 0) as likes,
                COALESCE(SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END), 0) as dislikes
            FROM likes
            WHERE post_id = ?
        `
    } else {
        query = `
            SELECT 
                COALESCE(SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END), 0) as likes,
                COALESCE(SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END), 0) as dislikes
            FROM likes
            WHERE comment_id = ?
        `
    }
    
    err = DB.QueryRow(query, targetID).Scan(&likes, &dislikes)
    return
}

func UpsertLike(userID, targetID int, isLike bool, isPost bool) error {
    tx, err := DB.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    var existingLike sql.NullBool
    var selectQuery, insertQuery, updateQuery, deleteQuery string

    if isPost {
        selectQuery = "SELECT is_like FROM likes WHERE user_id = ? AND post_id = ? AND comment_id IS NULL"
        insertQuery = "INSERT INTO likes (user_id, post_id, comment_id, is_like) VALUES (?, ?, NULL, ?)"
        updateQuery = "UPDATE likes SET is_like = ? WHERE user_id = ? AND post_id = ? AND comment_id IS NULL"
        deleteQuery = "DELETE FROM likes WHERE user_id = ? AND post_id = ? AND comment_id IS NULL"
    } else {
        selectQuery = "SELECT is_like FROM likes WHERE user_id = ? AND comment_id = ? AND post_id IS NULL"
        insertQuery = "INSERT INTO likes (user_id, post_id, comment_id, is_like) VALUES (?, NULL, ?, ?)"
        updateQuery = "UPDATE likes SET is_like = ? WHERE user_id = ? AND comment_id = ? AND post_id IS NULL"
        deleteQuery = "DELETE FROM likes WHERE user_id = ? AND comment_id = ? AND post_id IS NULL"
    }

    err = tx.QueryRow(selectQuery, userID, targetID).Scan(&existingLike)

    if err != nil && err != sql.ErrNoRows {
        return err
    }

    if err == sql.ErrNoRows {
        // No existing like, insert new one
        _, err = tx.Exec(insertQuery, userID, targetID, isLike)
    } else if existingLike.Valid {
        if existingLike.Bool == isLike {
            // User is toggling off their like/dislike
            _, err = tx.Exec(deleteQuery, userID, targetID)
        } else {
            // User is changing from like to dislike or vice versa
            _, err = tx.Exec(updateQuery, isLike, userID, targetID)
        }
    }

    if err != nil {
        return err
    }

    return tx.Commit()
}

func AddCreatedAtToLikesTable() error {
	_, err := DB.Exec(`
        ALTER TABLE likes ADD COLUMN created_at DATETIME DEFAULT CURRENT_TIMESTAMP;
    `)
	if err != nil {
		// If the error is because the column already exists, we can ignore it
		if err.Error() != "duplicate column name: created_at" {
			log.Printf("Error adding created_at column to likes table: %v", err)
			return err
		}
	}
	log.Println("created_at column added to likes table (if it didn't exist)")
	return nil
}

func GetPostsByCategory(categoryID int) ([]Post, error) {
    query := `
        SELECT DISTINCT p.id, p.title, p.content, u.username, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        JOIN post_categories pc ON p.id = pc.post_id
        WHERE pc.category_id = ?
        ORDER BY p.created_at DESC
    `
    return fetchPosts(query, categoryID)
}

func GetPostsByUser(userID int) ([]Post, error) {
    query := `
        SELECT p.id, p.title, p.content, u.username, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.user_id = ?
        ORDER BY p.created_at DESC
    `
    return fetchPosts(query, userID)
}

func GetLikedPostsByUser(userID int) ([]Post, error) {
    query := `
        SELECT p.id, p.title, p.content, u.username, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        JOIN likes l ON p.id = l.post_id
        WHERE l.user_id = ? AND l.is_like = 1
        ORDER BY p.created_at DESC
    `
    return fetchPosts(query, userID)
}

func fetchPosts(query string, args ...interface{}) ([]Post, error) {
    rows, err := DB.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var posts []Post
    for rows.Next() {
        var p Post
        err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.CreatedAt)
        if err != nil {
            return nil, err
        }
        posts = append(posts, p)
    }
    return posts, nil
}