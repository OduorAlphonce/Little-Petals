# Little Petals - MVP Issue Tracker

This document breaks down the 14-Day MVP Development Plan into actionable, GitHub-style issue templates. Developers can use these templates to populate their issue tracker (e.g., GitHub Issues, Jira, Trello).

---

## 📂 Table of Issues

- [Issue #1: Implement Configuration, DB Connection & SQL Migrations](#issue-1-implement-configuration-db-connection--sql-migrations)
- [Issue #2: Define User & Post Models and Repositories](#issue-2-define-user--post-models-and-repositories)
- [Issue #3: Secure Backend Auth Services & Session Middleware](#issue-3-secure-backend-auth-services--session-middleware)
- [Issue #4: Create User Auth Templates & Form Handlers](#issue-4-create-user-auth-templates--form-handlers)
- [Issue #5: Implement Blog & Poem Posting & Views](#issue-5-implement-blog--poem-posting--views)
- [Issue #6: Set Up Audio File Uploads & Custom Audio Player for Songs](#issue-6-set-up-audio-file-uploads--custom-audio-player-for-songs)
- [Issue #7: Build Home Feed & Creator Dashboard Pages](#issue-7-build-home-feed--creator-dashboard-pages)
- [Issue #8: Create Book & Chapter Data Layer](#issue-8-create-book--chapter-data-layer)
- [Issue #9: Build Book & Chapter Creator Interface](#issue-9-build-book--chapter-creator-interface)
- [Issue #10: Implement Multi-Chapter Reader & Navigation Pages](#issue-10-implement-multi-chapter-reader--navigation-pages)
- [Issue #11: Add Real-Time, Async Comment Threads via HTMX](#issue-11-add-real-time-async-comment-threads-via-htmx)
- [Issue #12: Build Reactive Reactions Bar (👍, ❤️, 👏)](#issue-12-build-reactive-reactions-bar---)
- [Issue #13: Implement End-To-End Authorization Checks & Input Validation](#issue-13-implement-end-to-end-authorization-checks--input-validation)
- [Issue #14: Styling Polish & Launch Script Deployment](#issue-14-styling-polish--launch-script-deployment)

---

### Issue #1: Implement Configuration, DB Connection & SQL Migrations
*   **Phase:** 1 (Day 1)
*   **Label:** `backend` `infrastructure`
*   **Estimate:** 4 hours

#### 📝 Description
We need to bootstrap the application's configuration loading system and verify database connectivity. PostgreSQL is chosen as the persistent storage layer. 

#### 🛠️ Tasks
- [ ] Implement [config.go](file:///home/aloduor/Little-Petals/internal/config/config.go) to read environment configurations (port, DB URI, static upload directories) from a local `.env` file.
- [ ] Implement database connection pooling inside [database.go](file:///home/aloduor/Little-Petals/internal/database/database.go) using standard SQL driver connections.
- [ ] Define DB tables in `web/migrations` files:
    - `001_create_users.sql` (schema: id, name, email, password_hash, created_at)
    - `002_create_posts.sql` (schema: id, user_id, title, description, content, post_type, file_url, created_at)
- [ ] Write a startup routine in [main.go](file:///home/aloduor/Little-Petals/cmd/server/main.go) to automatically execute pending SQL migrations.

#### ✅ Acceptance Criteria
- Running `go run cmd/server/main.go` connects to the DB without errors.
- Database tables `users` and `posts` are created in Postgres.

---

### Issue #2: Define User & Post Models and Repositories
*   **Phase:** 1 (Day 2)
*   **Label:** `backend` `data-layer`
*   **Estimate:** 6 hours

#### 📝 Description
Establish the Golang structs representing our database tables and write the CRUD interfaces (repositories) to fetch/persist data.

#### 🛠️ Tasks
- [ ] Define entity models in [user.go](file:///home/aloduor/Little-Petals/internal/models/user.go) and [post.go](file:///home/aloduor/Little-Petals/internal/models/post.go).
- [ ] Create repository interfaces and SQL implementations in [user_repo.go](file:///home/aloduor/Little-Petals/internal/repository/user_repo.go):
    - `Create(user *User)`
    - `FindByEmail(email string)`
    - `FindByID(id int)`
- [ ] Create repository interfaces and SQL implementations in [post_repo.go](file:///home/aloduor/Little-Petals/internal/repository/post_repo.go):
    - `Create(post *Post)`
    - `FindByID(id int)`
    - `ListAll(filterType string)`
    - `ListByUserID(userID int)`

#### ✅ Acceptance Criteria
- DB repositories successfully parse database results into Golang model structs.
- Code is covered by unit tests verifying retrieval and insert functions using a test database.

---

### Issue #3: Secure Backend Auth Services & Session Middleware
*   **Phase:** 2 (Day 3)
*   **Label:** `backend` `security`
*   **Estimate:** 5 hours

#### 📝 Description
Users must be authenticated to create content. Implement secure user signups, credentials validation, session persistence using HTTP cookies, and route protection.

#### 🛠️ Tasks
- [ ] Write logic in [auth_service.go](file:///home/aloduor/Little-Petals/internal/services/auth_service.go) to register users and verify login credentials.
- [ ] Use `bcrypt` library to hash passwords during signup and verify them during login.
- [ ] Write session serialization logic. Set cookie response headers containing session tokens.
- [ ] Implement Auth Middleware in [auth.go](file:///home/aloduor/Little-Petals/internal/middleware/auth.go) that decodes session tokens, loads the active user into context, and redirects unauthenticated requests.

#### ✅ Acceptance Criteria
- User passwords are saved encrypted inside the database.
- Private routes are restricted and redirect users to login.

---

### Issue #4: Create User Auth Templates & Form Handlers
*   **Phase:** 2 (Day 4)
*   **Label:** `frontend` `backend` `auth`
*   **Estimate:** 5 hours

#### 📝 Description
Build the client-side login and registration web pages and tie them to backend routing handlers.

#### 🛠️ Tasks
- [ ] Style user registration page inside [register.html](file:///home/aloduor/Little-Petals/web/templates/pages/register.html) using Tailwind CSS.
- [ ] Style login page inside [login.html](file:///home/aloduor/Little-Petals/web/templates/pages/login.html).
- [ ] Add handlers in [auth_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/auth_handler.go) parsing POST requests, triggering authentication, and setting cookie states.
- [ ] Display error alerts on forms (e.g. "Email already registered" or "Wrong password").

#### ✅ Acceptance Criteria
- Users can access the website, register an account, log in, and view protected pages.
- Logging out clears session cookies.

---

### Issue #5: Implement Blog & Poem Posting & Views
*   **Phase:** 3 (Day 5)
*   **Label:** `frontend` `backend` `content`
*   **Estimate:** 6 hours

#### 📝 Description
Creatives should be able to write and publish standard text blogs and structural poetry.

#### 🛠️ Tasks
- [ ] Wire routing in [post_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/post_handler.go) to handle creating posts of type `blog` and `poem`.
- [ ] Style the creation form in [create_post.html](file:///home/aloduor/Little-Petals/web/templates/pages/create_post.html) (add toggle styling to switch fields or parameters based on type).
- [ ] Design [view_post.html](file:///home/aloduor/Little-Petals/web/templates/pages/view_post.html). Apply CSS classes (e.g., `whitespace-pre-wrap` and serif typography) to preserve formatting line breaks for poetry.

#### ✅ Acceptance Criteria
- Users can write, publish, and view blogs and poems.
- Linebreaks and spaces in poetry are fully preserved when rendering the post.

---

### Issue #6: Set Up Audio File Uploads & Custom Audio Player for Songs
*   **Phase:** 3 (Day 6)
*   **Label:** `backend` `frontend` `media`
*   **Estimate:** 8 hours

#### 📝 Description
Musicians want to publish songs. Allow them to upload audio files (`.mp3` or `.wav`) and build a frontend media player for listeners.

#### 🛠️ Tasks
- [ ] Add secure file-upload parsing logic inside [post_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/post_handler.go).
- [ ] Store files inside `/web/static/uploads/songs/` and save file path references to the post's database row.
- [ ] Update [view_post.html](file:///home/aloduor/Little-Petals/web/templates/pages/view_post.html) to render an HTML5 audio container if `post_type == 'song'`.
- [ ] Add UX progress bar inside `/web/static/js/upload.js` displaying live upload statistics.

#### ✅ Acceptance Criteria
- Songs can be uploaded and stored locally.
- Page renders a stylized custom audio player that successfully plays the song.

---

### Issue #7: Build Home Feed & Creator Dashboard Pages
*   **Phase:** 3 (Day 7)
*   **Label:** `frontend` `backend` `dashboard`
*   **Estimate:** 6 hours

#### 📝 Description
Design the portal interfaces: a home feed listing all public posts and a personal creator dashboard.

#### 🛠️ Tasks
- [ ] Update [home.html](file:///home/aloduor/Little-Petals/web/templates/pages/home.html) to display latest creations utilizing [post_card.html](file:///home/aloduor/Little-Petals/web/templates/components/post_card.html) templates.
- [ ] Build tabs to filter the main feed by content type (All, Blogs, Poems, Songs).
- [ ] Design [dashboard.html](file:///home/aloduor/Little-Petals/web/templates/pages/dashboard.html) letting creators view their published pieces, draft count, and create entry buttons.

#### ✅ Acceptance Criteria
- Feeds render correct cards corresponding to the content type.
- Navigating tabs updates lists. Creators see their own items on their dashboard.

---

### Issue #8: Create Book & Chapter Data Layer
*   **Phase:** 4 (Day 8)
*   **Label:** `backend` `database`
*   **Estimate:** 6 hours

#### 📝 Description
Establish database storage architecture for nested books that can support multiple chapters.

#### 🛠️ Tasks
- [ ] Write SQL tables `003_create_books.sql` and `004_create_chapters.sql` containing unique foreign key constraints linking chapters sequentially to a book.
- [ ] Define schemas in [book.go](file:///home/aloduor/Little-Petals/internal/models/book.go) and [chapter.go](file:///home/aloduor/Little-Petals/internal/models/chapter.go).
- [ ] Write repository code in [book_repo.go](file:///home/aloduor/Little-Petals/internal/repository/book_repo.go):
    - `Create(book *Book)`
    - `AddChapter(chapter *Chapter)`
    - `GetBookWithChapters(bookID int)`
    - `GetChapterByNumber(bookID int, chapterNum int)`

#### ✅ Acceptance Criteria
- Migrations set up nested relationship rules.
- Repositories can insert chapters and fetch books containing chapter listings sorted by sequencing number.

---

### Issue #9: Build Book & Chapter Creator Interface
*   **Phase:** 4 (Day 9)
*   **Label:** `frontend` `backend` `content`
*   **Estimate:** 6 hours

#### 📝 Description
Build the creator dashboard forms to compile multi-chapter books.

#### 🛠️ Tasks
- [ ] Build [create_book.html](file:///home/aloduor/Little-Petals/web/templates/pages/create_book.html) with fields for book title, description, and cover image upload.
- [ ] Build a page / form interface to add a new chapter to an existing book (defining title, content, chapter number).
- [ ] Connect handlers in [book_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/book_handler.go) and [chapter_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/chapter_handler.go) to process requests.

#### ✅ Acceptance Criteria
- Users can create a book draft, upload custom cover art, and add subsequent chapters.
- Book cover files are saved correctly inside `/web/static/uploads/covers/`.

---

### Issue #10: Implement Multi-Chapter Reader & Navigation Pages
*   **Phase:** 4 (Day 10)
*   **Label:** `frontend` `navigation`
*   **Estimate:** 5 hours

#### 📝 Description
Create the landing index and reading screens for books.

#### 🛠️ Tasks
- [ ] Implement [view_book.html](file:///home/aloduor/Little-Petals/web/templates/pages/view_book.html) listing book details (metadata, author, cover) alongside the sequential table of contents.
- [ ] Build [chapter.html](file:///home/aloduor/Little-Petals/web/templates/pages/chapter.html) as a distraction-free, highly readable reading workspace.
- [ ] Implement "Next Chapter" and "Previous Chapter" pagination controls inside chapters.

#### ✅ Acceptance Criteria
- Readers can click on chapter listings on a book index to navigate directly.
- Reader pagination correctly fetches adjacent chapters.

---

### Issue #11: Add Real-Time, Async Comment Threads via HTMX
*   **Phase:** 5 (Day 11)
*   **Label:** `frontend` `backend` `htmx`
*   **Estimate:** 6 hours

#### 📝 Description
Readers should be able to comment on posts and chapters. We want comments to submit and display in real time without refreshing the entire page.

#### 🛠️ Tasks
- [ ] Write SQL migration `006_create_comments.sql` and corresponding model/repository files.
- [ ] Build comment POST route in [comment_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/comment_handler.go).
- [ ] Integrate HTMX on the comments input form: post content via ajax and append the returned HTML snippet directly to the DOM list.

#### ✅ Acceptance Criteria
- Submitting the comment form adds the comment instantly to the page without a full reload.
- Comments are persisted inside the DB.

---

### Issue #12: Build Reactive Reactions Bar (👍, ❤️, 👏)
*   **Phase:** 5 (Day 12)
*   **Label:** `frontend` `backend` `htmx`
*   **Estimate:** 7 hours

#### 📝 Description
Add engagement toggles allowing readers to react to content with reactions like "Like", "Love", or "Clap" asynchronously.

#### 🛠️ Tasks
- [ ] Populate schema `005_create_reactions.sql` tracking unique constraints (user ID, post ID, reaction type).
- [ ] Implement handler and repos in [reaction_handler.go](file:///home/aloduor/Little-Petals/internal/handlers/reaction_handler.go) and [reaction_repo.go](file:///home/aloduor/Little-Petals/internal/repository/reaction_repo.go).
- [ ] Design [reaction_bar.html](file:///home/aloduor/Little-Petals/web/templates/components/reaction_bar.html) containing reaction counts.
- [ ] Connect HTMX requests to toggle reactions: clicking a reaction increments/decrements active count dynamically.

#### ✅ Acceptance Criteria
- Clicking a reaction toggle updates counter stats instantly without screen flashes.
- Users can react and remove their reactions.

---

### Issue #13: Implement End-To-End Authorization Checks & Input Validation
*   **Phase:** 6 (Day 13)
*   **Label:** `security` `backend`
*   **Estimate:** 5 hours

#### 📝 Description
Protect backend controllers against unauthorized modifications and data corruption.

#### 🛠️ Tasks
- [ ] Implement route authorization checks ensuring users can only modify or delete content they authored.
- [ ] Sanitize user markdown text outputs to prevent scripting vulnerabilities (XSS).
- [ ] Validate file upload types: reject executable files and restrict files exceeding a maximum size (e.g., limit song uploads to 25MB).

#### ✅ Acceptance Criteria
- Trying to edit another user's post API directly throws 403 Forbidden.
- Attempting to upload non-media formats yields a clean validation error.

---

### Issue #14: Styling Polish & Launch Script Deployment
*   **Phase:** 6 (Day 14)
*   **Label:** `frontend` `devops`
*   **Estimate:** 5 hours

#### 📝 Description
Prepare the layout styling design to match a modern aesthetic and build quick execution commands.

#### 🛠️ Tasks
- [ ] Establish a warm, creative pink/rose layout color scheme using Tailwind variables inside [style.css](file:///home/aloduor/Little-Petals/web/static/css/style.css).
- [ ] Polish UI card grids, hover animations, and loading skeletons.
- [ ] Set up Nginx configuration templates and production execution scripts in `scripts/build.sh` and `scripts/run.sh`.

#### ✅ Acceptance Criteria
- Main interfaces look modern, styled, and display correctly on mobile sizes.
- Platform compiles cleanly using the build script and executes on target ports.
