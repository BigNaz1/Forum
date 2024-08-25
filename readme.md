# Reboot Forums

## Description

Reboot Forums is a web-based forum application that facilitates communication between users. It allows registered users to create posts, comment on posts, and interact through likes and dislikes. The forum also features category association for posts and filtering capabilities.

## Key Features

- User registration and authentication
- Creating and viewing posts
- Commenting on posts
- Liking and disliking posts and comments
- Associating categories with posts
- Filtering posts by categories, user-created posts, and liked posts
- Guest browsing (limited to viewing posts and comments)

## Technologies Used

- Go version 1.23
- SQLite database
- HTML
- CSS
- Minimal JavaScript

## Project Structure

The project follows a standard Go web application structure. Key components include:

- `main.go`: Entry point of the application
- `handlers/`: Contains HTTP request handlers
- `templates/`: HTML templates for rendering pages
- `static/`: Static assets (CSS)
- `forum.db`: SQLite database output file

## Authentication

Authentication in Reboot Forums is handled using session-based cookies with UUID tokens. The process includes:

1. **Registration**:
   - Users provide a username, email, and password.
   - The system checks for existing usernames or emails to prevent duplicates.
   - Passwords are hashed using bcrypt before storage in the database.
   - Upon successful registration, a session is created and a cookie is set.

2. **Login**:
   - Users enter their username and password.
   - The system verifies the credentials against the database.
   - If valid, any existing sessions for the user are deleted.
   - A new session is created with a UUID token, stored in the database and set as a cookie.

3. **Session Management**:
   - Sessions have a 24-hour expiration period.
   - Session tokens are generated using UUID for security.
   - Sessions are stored in the database, linking the token to the user ID.

4. **Logout**:
   - The session is deleted from the database.
   - The session cookie is cleared from the client.

5. **Security Measures**:
   - Passwords are hashed using bcrypt for secure storage.
   - Session cookies are HTTP-only and secure (when using HTTPS) to prevent XSS attacks.
   - The system uses prepared statements to prevent SQL injection.

This authentication system ensures secure user registration, login, and session management, protecting user data and preventing unauthorized access.

## Database

Reboot Forums uses SQLite as its database system. The database structure and operations are managed through a custom Go package.

### Database Schema

The database consists of the following tables:

1. `users`: Stores user information (id, username, email, password).
2. `posts`: Contains all forum posts (id, user_id, title, content, created_at, updated_at).
3. `comments`: Stores comments on posts (id, post_id, user_id, content, created_at, updated_at).
4. `categories`: Defines post categories (id, name).
5. `post_categories`: Links posts to categories (post_id, category_id).
6. `likes`: Tracks likes and dislikes for posts and comments (id, user_id, post_id, comment_id, is_like, created_at).
7. `sessions`: Manages user sessions (id, user_id, token, expiry, is_guest, last_activity, created_at).

### Key Database Operations

- **Initialization**: The database connection is established using the `InitDB` function.
- **Table Creation**: Tables are created if they don't exist using the `CreateTables` function.
- **Default Categories**: A set of default categories is added to the database on initialization.
- **Like System**: The database supports a comprehensive like/dislike system for both posts and comments.
- **Post Retrieval**: Functions are available to fetch posts by category, user, or liked posts.
- **Transaction Support**: The like system uses transactions to ensure data integrity.

### Notable Features

- Use of prepared statements to prevent SQL injection.
- Automatic timestamp management for creation and update times.
- Support for guest sessions.
- Efficient querying with JOINs for related data retrieval.

This database structure allows for efficient storage and retrieval of forum data, supporting all the key features of the Reboot Forums application.

## Post Handling System

The Reboot Forums post handling system manages the creation, viewing, editing, and deletion of posts, as well as handling likes and comments. Here's a detailed overview of its functionalities:

### Post Creation

- **Handler**: `CreatePostFormHandler`
- **Features**:
  - Displays a form for creating new posts (GET request)
  - Processes the form submission to create a new post (POST request)
  - Validates user authentication before allowing post creation
  - Supports associating multiple categories with a post
  - Uses database transactions to ensure data integrity when creating posts

### Viewing Posts

- **Handler**: `ViewPostHandler`
- **Features**:
  - Retrieves and displays a single post along with its details
  - Fetches associated categories and comments for the post
  - Handles cases where the post doesn't exist
  - Determines if the current user is the author of the post

### Liking Posts and Comments

- **Handlers**: `LikePostHandler`, `LikeCommentHandler`
- **Features**:
  - Allows authenticated users to like or dislike posts and comments
  - Uses a flexible system that can toggle between like and dislike
  - Updates like/dislike counts in real-time
  - Returns updated counts as JSON for AJAX requests

### Deleting Posts

- **Handler**: `DeletePostHandler`
- **Features**:
  - Allows post authors to delete their posts
  - Validates user authorization before allowing deletion
  - Implements cascading deletion for associated data (categories, likes, comments)
  - Uses database transactions to ensure all related data is deleted consistently

### Helper Functions

- `getPost`: Retrieves detailed post information, including like counts
- `getPostCategories`: Fetches categories associated with a post
- `updatePost`: Handles the database operations for updating a post
- `deletePost`: Manages the database operations for deleting a post and its associated data

### Security Measures

- User authentication is required for creating, editing, liking, and deleting posts
- Authorization checks ensure users can only edit or delete their own posts
- Input validation is performed to prevent invalid data submission
- Database transactions are used to maintain data integrity during complex operations

### Database Interactions

- Utilizes prepared statements to prevent SQL injection
- Implements efficient querying, including JOIN operations for fetching related data
- Handles potential database errors and provides appropriate error responses

This post handling system provides a robust and secure way to manage forum posts, ensuring that users can interact with content while maintaining data integrity and user permissions.

## Comment Handling System

The Reboot Forums comment handling system manages the creation and retrieval of comments for posts. It also integrates with the like/dislike functionality for comments. Here's a detailed overview of its functionalities:

### Retrieving Comments

- **Function**: `getCommentsByPostID`
- **Features**:
  - Fetches all comments for a specific post
  - Retrieves comment details including ID, content, author, and creation time
  - Orders comments chronologically (oldest first)
  - Integrates with the like system to fetch like/dislike counts for each comment
  - Uses JOIN operation for efficient data retrieval

### Adding Comments

- **Handler**: `AddCommentHandler`
- **Features**:
  - Allows authenticated users to add comments to posts
  - Validates user authentication before allowing comment submission
  - Checks for valid post ID and non-empty comment content
  - Redirects user back to the post page after successful comment addition

### Comment Creation

- **Function**: `addComment`
- **Features**:
  - Inserts a new comment into the database
  - Associates the comment with the user and the post
  - Records the creation timestamp

### Utility Functions

- `getPostIDFromCommentID`: Retrieves the post ID associated with a given comment ID

### Security Measures

- User authentication is required for adding comments
- Input validation ensures non-empty comment content and valid post IDs
- Uses prepared statements for database queries to prevent SQL injection

### Database Interactions

- Efficient querying using JOIN operations for fetching comment data
- Handles potential database errors and provides appropriate error responses
- Integrates with the like system to provide comprehensive comment data

### Integration with Like System

- Fetches like and dislike counts for each comment
- Allows for displaying popularity of comments alongside their content

This comment handling system provides a robust way to manage and display comments on forum posts. It ensures that only authenticated users can add comments, maintains data integrity, and integrates seamlessly with the post and like systems of the forum.

## Setup and Usage

1. Ensure Go 1.23 is installed on your system.
2. Clone the repository.
3. Navigate to the project directory.
4. Run the application:

Or use the provided Dockerfile to build and run the application in a container.

5. Access the forum through a web browser at `http://localhost:8080` (or the appropriate port).

## Docker Support

Reboot Forums includes Docker support for easy deployment and consistent environments across different systems.

### Dockerfile

The project includes a Dockerfile that sets up the necessary environment for running the application. Key features of the Dockerfile include:

- Base image: `golang:1.23-alpine`
- Installation of required packages:
  - `sqlite` and `sqlite-dev` for database support
  - `gcc` and `musl-dev` for compilation of C dependencies
  - `git` for potential version control operations
  - `curl` for network utility
  - `tzdata` for timezone data
  - `ca-certificates` for secure connections

### Building and Running with Docker

To build and run the application using Docker:

1. Build the Docker image:
   ```
   docker build -t reboot-forums:latest .
   ```

2. Run the Docker container:
   ```
   docker run -d --name reboot-forums -p 8080:8080 reboot-forums:latest
   ```

### Docker Compose (Optional)

For easier management of the application, especially if additional services are added in the future, a `docker-compose.yml` file can be used. This file is not currently included in the project but can be easily added to orchestrate the application and any additional services.

### Benefits of Docker Usage

- Consistent environment across development, testing, and production
- Easy deployment and scaling
- Isolation of the application and its dependencies
- Simplified setup process for new developers joining the project

### Note on Database Persistence

When using Docker, be aware that the SQLite database file is created inside the container. For data persistence between container restarts, consider using a Docker volume to store the database file.


## Features in Detail

- **Post Creation**: Registered users can create posts and associate them with one or more categories.
- **Commenting**: Registered users can comment on posts.
- **Likes and Dislikes**: Registered users can like or dislike posts and comments.
- **Filtering**: Users can filter posts by categories. Registered users can also filter by their created posts or liked posts.
- **User Profiles**: Each user has a profile showcasing their posts and activity.

## License

This project is developed by:
- Nezar Jaberi
- Abdul Aziz Bin Rajab
- Mobeen Zakir
- Mahmood Abdulla

## Note

This project was developed as part of an educational assignment. It demonstrates practical application of web development concepts using Go and database management with SQLite.