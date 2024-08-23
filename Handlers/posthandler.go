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
		Error404Handler(w, r)
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
		Error500Handler(w, r)
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

	err = RenderTemplate(w, "create-post.html", data)
	if err != nil {
		log.Printf("Error rendering create-post template: %v", err)
		Error500Handler(w, r)
		return
	}
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		Error400Handler(w, r)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	categoryIDs := r.Form["categories"]

	if title == "" || content == "" {
		Error400Handler(w, r)
		return
	}

	if len(title) == 0 || len(title) > MaxTitleLength {
		Error400Handler(w, r)
		return
	}

	if len(content) == 0 || len(content) > MaxPostLength {
		Error400Handler(w, r)
		return
	}

	categories := make([]int, 0, len(categoryIDs))
	for _, id := range categoryIDs {
		catID, err := strconv.Atoi(id)
		if err != nil {
			Error400Handler(w, r)
			return
		}
		categories = append(categories, catID)
	}

	postID, err := createPost(user.ID, title, content, categories)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		Error500Handler(w, r)
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
		Error400Handler(w, r)
		return
	}

	post, err := getPost(postID)
	if err == sql.ErrNoRows {
		Error404Handler(w, r)
		return
	} else if err != nil {
		log.Printf("Error fetching post: %v", err)
		Error500Handler(w, r)
		return
	}

	categories, err := getPostCategories(postID)
	if err != nil {
		log.Printf("Error fetching post categories: %v", err)
		Error500Handler(w, r)
		return
	}

	comments, err := getCommentsByPostID(postID)
	if err != nil {
		log.Printf("Error fetching comments: %v", err)
		comments = []Comment{}
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

	err = RenderTemplate(w, "view-post.html", data)
	if err != nil {
		log.Printf("Error rendering view-post template: %v", err)
		Error500Handler(w, r)
		return
	}
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
		Error404Handler(w, r)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		Error400Handler(w, r)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		Error400Handler(w, r)
		return
	}

	isLike, err := strconv.ParseBool(r.FormValue("is_like"))
	if err != nil {
		Error400Handler(w, r)
		return
	}

	err = UpsertLike(user.ID, postID, isLike, true)
	if err != nil {
		log.Printf("Error upserting like: %v", err)
		Error500Handler(w, r)
		return
	}

	likes, dislikes, err := GetLikeCounts(postID, true)
	if err != nil {
		log.Printf("Error getting like counts: %v", err)
		Error500Handler(w, r)
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
		Error404Handler(w, r)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		Error400Handler(w, r)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	if err != nil {
		Error400Handler(w, r)
		return
	}

	isLike, err := strconv.ParseBool(r.FormValue("is_like"))
	if err != nil {
		Error400Handler(w, r)
		return
	}

	err = UpsertLike(user.ID, commentID, isLike, false)
	if err != nil {
		log.Printf("Error upserting comment like: %v", err)
		Error500Handler(w, r)
		return
	}

	likes, dislikes, err := GetLikeCounts(commentID, false)
	if err != nil {
		log.Printf("Error getting comment like counts: %v", err)
		Error500Handler(w, r)
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
		Error404Handler(w, r)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		Error400Handler(w, r)
		return
	}

	postID, err := strconv.Atoi(r.URL.Path[len("/delete-post/"):])
	if err != nil {
		Error400Handler(w, r)
		return
	}

	var authorID int
	err = DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&authorID)
	if err != nil {
		log.Printf("Error fetching post author: %v", err)
		Error500Handler(w, r)
		return
	}

	if authorID != user.ID {
		Error500Handler(w, r)
		return
	}

	err = deletePost(postID)
	if err != nil {
		log.Printf("Error deleting post: %v", err)
		Error500Handler(w, r)
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

	_, err = tx.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
