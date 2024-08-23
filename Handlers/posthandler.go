package RebootForums

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

func CreatePostFormHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		displayCreatePostForm(w, r)
	case http.MethodPost:
		handleCreatePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func displayCreatePostForm(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	categories, err := GetAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username   string
		Categories []Category
		LoggedIn   bool
	}{
		Username:   user.Username,
		Categories: categories,
		LoggedIn:   true,
	}

	RenderTemplate(w, "create-post.html", data)
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "You must be logged in to create a post", http.StatusUnauthorized)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	categoryIDs := r.Form["categories"]

	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	categories := make([]int, 0, len(categoryIDs))
	for _, id := range categoryIDs {
		catID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
		categories = append(categories, catID)
	}

	postID, err := createPost(user.ID, title, content, categories)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

func createPost(userID int, title, content string, categories []int) (int, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
        INSERT INTO posts (user_id, title, content, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)
    `, userID, title, content, time.Now(), time.Now())
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, categoryID := range categories {
		_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
		if err != nil {
			return 0, err
		}
	}

	return int(postID), tx.Commit()
}

func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.URL.Path[len("/post/"):])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := getPost(postID)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		log.Printf("Error fetching post: %v", err)
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		return
	}

	categories, err := getPostCategories(postID)
	if err != nil {
		log.Printf("Error fetching post categories: %v", err)
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		return
	}

	comments, err := getCommentsByPostID(postID)
	if err != nil {
		log.Printf("Error fetching comments: %v", err)
		comments = []Comment{} // Use empty slice instead of nil
	}

	user, err := GetUserFromSession(r)
	loggedIn := err == nil && user != nil
	var username string
	var isAuthor bool

	if loggedIn {
		username = user.Username
		isAuthor = user.Username == post.Author
	}

	data := struct {
		Post       Post
		Categories []string
		Comments   []Comment
		IsAuthor   bool
		LoggedIn   bool
		Username   string
	}{
		Post:       post,
		Categories: categories,
		Comments:   comments,
		IsAuthor:   isAuthor,
		LoggedIn:   loggedIn,
		Username:   username,
	}

	RenderTemplate(w, "view-post.html", data)
}

func getPost(postID int) (Post, error) {
	var post Post
	var likes, dislikes sql.NullInt64

	err := DB.QueryRow(`
        SELECT p.id, p.title, p.content, u.username, p.created_at,
               COALESCE(l.likes, 0) as likes, COALESCE(l.dislikes, 0) as dislikes
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN (
            SELECT post_id,
                   SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END) as likes,
                   SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END) as dislikes
            FROM likes
            WHERE post_id = ?
            GROUP BY post_id
        ) l ON p.id = l.post_id
        WHERE p.id = ?
    `, postID, postID).Scan(
		&post.ID, &post.Title, &post.Content, &post.Author, &post.CreatedAt,
		&likes, &dislikes,
	)

	if err != nil {
		return post, err
	}

	post.Likes = int(likes.Int64)
	post.Dislikes = int(dislikes.Int64)

	return post, nil
}

func getPostCategories(postID int) ([]string, error) {
	rows, err := DB.Query(`
        SELECT c.name
        FROM categories c
        JOIN post_categories pc ON c.id = pc.category_id
        WHERE pc.post_id = ?
    `, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "You must be logged in to like/dislike posts", http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	isLike, err := strconv.ParseBool(r.FormValue("is_like"))
	if err != nil {
		http.Error(w, "Invalid like value", http.StatusBadRequest)
		return
	}

	err = UpsertLike(user.ID, postID, isLike, true) // true indicates it's a post like
	if err != nil {
		log.Printf("Error upserting like: %v", err)
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	likes, dislikes, err := GetLikeCounts(postID, true) // true indicates it's a post
	if err != nil {
		log.Printf("Error getting like counts: %v", err)
		http.Error(w, "Error getting like counts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	})
}

func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "You must be logged in to like/dislike comments", http.StatusUnauthorized)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	isLike, err := strconv.ParseBool(r.FormValue("is_like"))
	if err != nil {
		http.Error(w, "Invalid like value", http.StatusBadRequest)
		return
	}

	err = UpsertLike(user.ID, commentID, isLike, false) // false indicates it's a comment like
	if err != nil {
		log.Printf("Error upserting comment like: %v", err)
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	likes, dislikes, err := GetLikeCounts(commentID, false) // false indicates it's a comment
	if err != nil {
		log.Printf("Error getting comment like counts: %v", err)
		http.Error(w, "Error getting like counts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	})
}

func updatePost(postID int, title, content string, categories []int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE posts SET title = ?, content = ?, updated_at = ? WHERE id = ?",
		title, content, time.Now(), postID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM post_categories WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	for _, categoryID := range categories {
		_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "You must be logged in to delete a post", http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(r.URL.Path[len("/delete-post/"):])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Check if the user is the author of the post
	var authorID int
	err = DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&authorID)
	if err != nil {
		log.Printf("Error fetching post author: %v", err)
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	if authorID != user.ID {
		http.Error(w, "You are not authorized to delete this post", http.StatusForbidden)
		return
	}

	err = deletePost(postID)
	if err != nil {
		log.Printf("Error deleting post: %v", err)
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deletePost(postID int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete associated records
	_, err = tx.Exec("DELETE FROM post_categories WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM likes WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM comments WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	// Delete the post
	_, err = tx.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
