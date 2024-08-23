package RebootForums

import (
    "log"
    "net/http"
    "strconv"
    "time"
)

func getCommentsByPostID(postID int) ([]Comment, error) {
    rows, err := DB.Query(`
        SELECT c.id, c.content, u.username, c.created_at
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC
    `, postID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var comments []Comment
    for rows.Next() {
        var comment Comment
        if err := rows.Scan(&comment.ID, &comment.Content, &comment.Author, &comment.CreatedAt); err != nil {
            return nil, err
        }
        // Get like counts for each comment
        comment.Likes, comment.Dislikes, err = GetLikeCounts(comment.ID, false)  // false indicates it's a comment
        if err != nil {
            return nil, err
        }
        comments = append(comments, comment)
    }
    return comments, nil
}

func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    user, err := GetUserFromSession(r)
    if err != nil {
        http.Error(w, "You must be logged in to comment", http.StatusUnauthorized)
        return
    }

    postID, err := strconv.Atoi(r.FormValue("post_id"))
    if err != nil {
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }

    content := r.FormValue("content")
    if content == "" {
        http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
        return
    }

    err = addComment(user.ID, postID, content)
    if err != nil {
        log.Printf("Error adding comment: %v", err)
        http.Error(w, "Error adding comment", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

func addComment(userID, postID int, content string) error {
    _, err := DB.Exec(`
        INSERT INTO comments (user_id, post_id, content, created_at)
        VALUES (?, ?, ?, ?)
    `, userID, postID, content, time.Now())
    return err
}

func getPostIDFromCommentID(commentID int) (int, error) {
    var postID int
    err := DB.QueryRow("SELECT post_id FROM comments WHERE id = ?", commentID).Scan(&postID)
    if err != nil {
        return 0, err
    }
    return postID, nil
}