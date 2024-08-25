package RebootForums

import "time"

const (
    MaxTitleLength    = 80
    MaxPostLength     = 3000
    MaxCommentLength  = 600
)

// Post represents a forum post
type Post struct {
    ID        int
    Title     string
    Content   string
    Author    string
    CreatedAt time.Time
    Likes     int
    Dislikes  int
}

func (p Post) FormattedCreatedAt() string {
    return p.CreatedAt.Format("January 2, 2006 at 3:04 PM")
}

// Comment represents a comment on a post
type Comment struct {
    ID        int
    PostID    int
    Content   string
    Author    string
    CreatedAt time.Time
    Likes     int
    Dislikes  int
}

// Category represents a forum category
type Category struct {
    ID   int
    Name string
}

// User represents a forum user
type User struct {
    ID       int
    Username string
    Email    string
    Password string
}

// GetAllCategories fetches all categories from the database
func GetAllCategories() ([]Category, error) {
    rows, err := DB.Query("SELECT id, name FROM categories ORDER BY name")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var categories []Category
    for rows.Next() {
        var c Category
        err := rows.Scan(&c.ID, &c.Name)
        if err != nil {
            return nil, err
        }
        categories = append(categories, c)
    }
    return categories, nil
}

