# Little Petals 14-Day MVP Development Plan

Little Petals is a content-focused social platform for creatives to publish books, songs, poetry, and blogs, and interact through comments and reactions. This plan outlines a realistic 14-day roadmap to build the Minimum Viable Product (MVP) from the existing skeleton codebase.

---

## 🏗️ Technology Stack & Scope

*   **Backend:** Go (Golang) 1.24+, `net/http` with the Chi Router.
*   **Database:** PostgreSQL with standard SQL migrations (`database/sql` driver).
*   **Frontend:** HTML5, Tailwind CSS, Vanilla JS, and **HTMX** (for reactive components).
*   **Template Engine:** Go `html/template` package.
*   **Assets:** Local storage for audio files (`songs/`), images (`covers/`), and books.

---

## 📅 The 14-Day Roadmap

```mermaid
gantt
    title Little Petals 14-Day MVP Timeline
    dateFormat  YYYY-MM-DD
    axisFormat  Day %e
    section Phase 1
    DB & Models Setup       :active, 2026-08-05, 2d
    section Phase 2
    Auth & Middleware      :2026-08-07, 2d
    section Phase 3
    Blogs, Poems, & Songs   :2026-08-09, 3d
    section Phase 4
    Multi-Chapter Books    :2026-08-12, 3d
    section Phase 5
    Reactions & Comments    :2026-08-15, 2d
    section Phase 6
    Polish & Launch Prep    :2026-08-17, 2d
```

### 🔹 Phase 1: Database & Models (Days 1 - 2)
**Goal:** Establish database connectivity and build database tables and basic database operations.

#### Day 1: Configuration, DB Connection & Migration CLI
*   **Tasks:**
    *   Implement [config.go](file:///home/aloduor/Little-Petals/internal/config/config.go) to read environmental variables (`.env`) for server port, database connection string, and asset paths.
    *   Set up connection pooling in [database.go](file:///home/aloduor/Little-Petals/internal/database/database.go) using `pgx` or `pq` drivers.
    *   Populate SQL migrations inside [web/migrations/](file:///home/aloduor/Little-Petals/web/migrations/):
        *   `001_create_users.sql` (IDs, names, email, hashed passwords, roles, timestamps).
        *   `002_create_posts.sql` (IDs, user_id, title, description, content, post_type [blog/poem/song], file_url, timestamps).
    *   Write a basic migration runner command inside `/cmd/server/main.go` or a separate helper function to apply migrations on startup.
*   **Verification:** Verify connection to PostgreSQL and confirm database tables are successfully created on startup.

#### Day 2: Models & Core Repositories (Users & Posts)
*   **Tasks:**
    *   Write the struct schemas in [user.go](file:///home/aloduor/Little-Petals/internal/models/user.go) and [post.go](file:///home/aloduor/Little-Petals/internal/models/post.go).
    *   Implement `UserRepository` interface and SQL implementations in [user_repo.go](file:///home/aloduor/Little-Petals/internal/repository/user_repo.go) (Create, FindByEmail, FindByID).
    *   Implement `PostRepository` interface and SQL implementations in [post_repo.go](file:///home/aloduor/Little-Petals/internal/repository/post_repo.go) (Create, FindByID, ListAll, ListByUserID).
*   **Verification:** Run basic unit tests inserting dummy users and posts to verify SQL queries.

---

### 🔹 Phase 2: Authentication & Sessions (Days 3 - 4)
**Goal:** Secure the platform and track logged-in users.

#### Day 3: Backend Auth Logic & Middleware
*   **Tasks:**
    *   Build authentication services in [auth_service.go](file:///home/aloduor/Little-Petals/internal/services/auth_service.go) (Register, Authenticate/Login). Use `bcrypt` for password security.
    *   Implement cookie-based session management. Store session tokens in memory or database.
    *   Create authentication and routing middleware in [auth.go](file:///home/aloduor/Little-Petals/internal/middleware/auth.go) to inject session context into requests.
*   **Verification:** Run unit tests for successful/failed authentication logic and password verification.

#### Day 4: Registration, Login & Logout Views
*   **Tasks:**
    *   Flesh out frontend pages in [register.html](file:///home/aloduor/Little-Petals/web/templates/pages/register.html) and [login.html](file:///home/aloduor/Little-Petals/web/templates/pages/login.html). Set up Tailwind styling.
    *   Implement form parsing and handler logic in [auth_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/auth_handler.go).
    *   Establish visual feedback in forms for invalid login credentials or validation errors.
*   **Verification:** Run server and test signup, login, cookie persistence, and logout flow using a browser.

---

### 🔹 Phase 3: Blogs, Poems & Songs (Days 5 - 7)
**Goal:** Allow creatives to publish content, write text blogs/poems, and upload audio files (songs).

#### Day 5: Blogs & Poems Creation
*   **Tasks:**
    *   Connect the `PostRepository` to the handlers in [post_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/post_handler.go).
    *   Update [create_post.html](file:///home/aloduor/Little-Petals/web/templates/pages/create_post.html): style inputs with Tailwind and build a toggle between Blogs (rich description) and Poems (formatted linebreaks/white-space).
    *   Write the view template in [view_post.html](file:///home/aloduor/Little-Petals/web/templates/pages/view_post.html) using CSS classes like `whitespace-pre-line` to handle linebreaks correctly for poetry.
*   **Verification:** Authenticate as a user, submit blogs and poems, and review their typography.

#### Day 6: Songs Upload & Custom Audio Player
*   **Tasks:**
    *   Implement file uploading logic in [post_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/post_handler.go) for `.mp3`/`.wav` songs. Save files locally in `/web/static/uploads/songs/`.
    *   Update [view_post.html](file:///home/aloduor/Little-Petals/web/templates/pages/view_post.html) to render a modern audio player component when viewing a post of type `song`.
    *   Inject styles into `/web/static/js/upload.js` to show upload progress indicators for larger audio files.
*   **Verification:** Upload dummy audio files, inspect storage folder path, and play audio in browser.

#### Day 7: Feed & Dashboard Integration
*   **Tasks:**
    *   Build home feed page in [home.html](file:///home/aloduor/Little-Petals/web/templates/pages/home.html) using [post_card.html](file:///home/aloduor/Little-Petals/web/templates/components/post_card.html) elements. Feed should display latest posts with filter tabs (All, Blogs, Poems, Songs).
    *   Build the creator dashboard in [dashboard.html](file:///home/aloduor/Little-Petals/web/templates/pages/dashboard.html) where users can see list of posts they published, view counts, and editing routes.
*   **Verification:** Click between dashboard and feed, confirming listings load properly.

---

### 🔹 Phase 4: Multi-Chapter Books (Days 8 - 10)
**Goal:** Implement books featuring multiple chapters, along with title covers and chapter indexes.

#### Day 8: Book & Chapter Migrations and Repositories
*   **Tasks:**
    *   Define SQL tables in `003_create_books.sql` and `004_create_chapters.sql`.
    *   Write schemas in [book.go](file:///home/aloduor/Little-Petals/internal/models/book.go) and [chapter.go](file:///home/aloduor/Little-Petals/internal/models/chapter.go).
    *   Implement SQL queries in [book_repo.go](file:///home/aloduor/Little-Petals/internal/repository/book_repo.go) (Create Book, Add Chapter, Get Book by ID with all chapters, Get Chapter by Number).
*   **Verification:** Confirm migration tables load and repositories fetch nested chapter sequences correctly.

#### Day 9: Creator Interfaces: Book & Chapter Creation
*   **Tasks:**
    *   Build [create_book.html](file:///home/aloduor/Little-Petals/web/templates/pages/create_book.html) containing forms for title, description, and cover image upload (saved in `/web/static/uploads/covers/`).
    *   Build handlers in [book_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/book_handler.go) and [chapter_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/chapter_handler.go) to support creating a book first, then clicking "Add Chapter" to insert sequential chapters.
*   **Verification:** Publish a book and append three test chapters sequentially. Check if covers are saved in the correct uploads folder.

#### Day 10: Book Reading Experience
*   **Tasks:**
    *   Create [view_book.html](file:///home/aloduor/Little-Petals/web/templates/pages/view_book.html) as a summary index displaying the book cover, description, author name, and a listed Table of Contents.
    *   Design [chapter.html](file:///home/aloduor/Little-Petals/web/templates/pages/chapter.html) as a clean, distraction-free reading view with "Previous Chapter" and "Next Chapter" navigation.
*   **Verification:** Navigate book index, select a chapter, and step back and forth using pagination controls.

---

### 🔹 Phase 5: Reactions & Comments (Days 11 - 12)
**Goal:** Enable engagement using HTMX-powered reactions and nested comments.

#### Day 11: Real-time Comments with HTMX
*   **Tasks:**
    *   Populate `006_create_comments.sql`. Write [comment.go](file:///home/aloduor/Little-Petals/internal/models/comment.go) schema and [comment_repo.go](file:///home/aloduor/Little-Petals/internal/repository/comment_repo.go).
    *   Establish handlers in [comment_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/comment_handler.go) that accept HTMX POST requests, persist the comments, and return HTML snippets of the comment.
    *   Update post and chapter templates to include comment input boxes. Submit comments asynchronously without full page reloads.
*   **Verification:** Post multiple comments on a blog, confirming they render immediately in the comments feed.

#### Day 12: Creative Reactions Bar
*   **Tasks:**
    *   Populate `005_create_reactions.sql` and implement repository layer [reaction_repo.go](file:///home/aloduor/Little-Petals/internal/repository/reaction_repo.go).
    *   Design [reaction_bar.html](file:///home/aloduor/Little-Petals/web/templates/components/reaction_bar.html) displaying count statistics for Like (👍), Love (❤️), and Clap (👏).
    *   Connect [reactions.js](file:///home/aloduor/Little-Petals/web/static/js/reactions.js) or HTMX endpoints in [reaction_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/reaction_handler.go) to toggle active reactions and update counts dynamically.
*   **Verification:** Press reaction buttons on a post and verify counts increment and decrement on click without reloads.

---

### 🔹 Phase 6: Polish & Launch (Days 13 - 14)
**Goal:** Secure, visual, and deployment-ready build.

#### Day 13: Edge Case Testing, Security & Validation Check
*   **Tasks:**
    *   Validate file uploads: restrict non-audio uploads in songs and block large files.
    *   Implement user authorization checks (e.g. users should not edit other users' posts/books).
    *   Sanitize text outputs to prevent Cross-Site Scripting (XSS) when rendering HTML templates.
    *   Validate responsive CSS across mobile screen widths.
*   **Verification:** Try accessing editing endpoints of another user's post directly to confirm access is forbidden.

#### Day 14: Styling Polish & Launch Script
*   **Tasks:**
    *   Enhance visual assets and stylesheets ([style.css](file:///home/aloduor/Little-Petals/web/static/css/style.css)) using a premium, warm creative palette (rose/pink tones fitting the "Petals" brand).
    *   Configure production configurations and create build / run scripts inside [scripts/](file:///home/aloduor/Little-Petals/scripts/).
    *   Draft final walkthrough and run performance sanity check.
*   **Verification:** Verify successful execution of `run.sh` script and compile production binary with `build.sh`.

---

## 💾 Core Database Schema Reference

```sql
-- Users Table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Posts Table (Blogs, Poems, Songs)
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT,
    post_type VARCHAR(50) CHECK (post_type IN ('blog', 'poem', 'song')),
    file_url VARCHAR(550), -- Used for songs (audio URL)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Books Table
CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_url VARCHAR(550),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Chapters Table
CREATE TABLE chapters (
    id SERIAL PRIMARY KEY,
    book_id INTEGER REFERENCES books(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    chapter_number INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(book_id, chapter_number)
);

-- Reactions Table
CREATE TABLE reactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
    reaction_type VARCHAR(50) NOT NULL, -- e.g., 'like', 'love', 'clap'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, post_id, reaction_type)
);

-- Comments Table
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🚀 Running Locally

To build and run the service locally, run these commands:

1.  **Configure Environment:**
    Set up database connection string and port in a `.env` file in the root directory:
    ```ini
    PORT=8080
    DB_CONN_STR="postgres://username:password@localhost:5432/littlepetals?sslmode=disable"
    ```
2.  **Run Development Server:**
    ```bash
    chmod +x scripts/run.sh
    ./scripts/run.sh
    ```
