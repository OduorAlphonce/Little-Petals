# Little-Petals
Little-Petals is a minimal content-focused social platform built in Go where creatives can publish books, music, poetry and blogs, and interact through reactions.

## uses

The app/website is able to accomodate the following use cases :

📚 Books (multi-chapter)

❤️ Reactions (like, love, clap, etc.)

📝 Blogs

🎵 Songs

✍ Poems

## Project Achitecture

The project follows a clean architecture separating handlers, services, and repositories to keep the backend modular and scalable.

```text
little-petals/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│
│   ├── config/
│   │   └── config.go
│   │
│   ├── database/
│   │   └── db.go
│   │
│   ├── models/
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── book.go
│   │   ├── chapter.go
│   │   ├── reaction.go
│   │   └── comment.go
│   │
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── post_handler.go
│   │   ├── book_handler.go
│   │   ├── chapter_handler.go
│   │   ├── reaction_handler.go
│   │   └── comment_handler.go
│   │
│   ├── middleware/
│   │   ├── auth.go
│   │   └── logger.go
│   │
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── post_repo.go
│   │   ├── book_repo.go
│   │   ├── reaction_repo.go
│   │   └── comment_repo.go
│   │
│   └── services/
│       ├── auth_service.go
│       ├── post_service.go
│       └── reaction_service.go
│
├── web/
│
│   ├── templates/
│   │
│   │   ├── layouts/
│   │   │   └── base.html
│   │   │
│   │   ├── pages/
│   │   │   ├── home.html
│   │   │   ├── login.html
│   │   │   ├── register.html
│   │   │   ├── dashboard.html
│   │   │   ├── create_post.html
│   │   │   ├── view_post.html
│   │   │   ├── create_book.html
│   │   │   ├── view_book.html
│   │   │   └── chapter.html
│   │   │
│   │   └── components/
│   │       ├── navbar.html
│   │       ├── post_card.html
│   │       └── reaction_bar.html
│
│   ├── static/
│   │
│   │   ├── css/
│   │   │   └── style.css
│   │   │
│   │   ├── js/
│   │   │   ├── main.js
│   │   │   ├── reactions.js
│   │   │   └── upload.js
│   │   │
│   │   ├── images/
│   │   │   └── logo.png
│   │   │
│   │   └── uploads/
│   │       ├── songs/
│   │       ├── covers/
│   │       └── books/
│
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_posts.sql
│   ├── 003_create_books.sql
│   ├── 004_create_chapters.sql
│   ├── 005_create_reactions.sql
│   └── 006_create_comments.sql
│
├── scripts/
│   ├── build.sh
│   └── run.sh
│
├── .env
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Tech Stack 

<mark>FrontEnd</mark>

HTML5

CSS3

TailWind CSS

JavaScript(Vanilla JS)

<mark>Optional FrontEnd helper</mark>

HTMX

<mark>BackEnd</mark>

Golang

Chi(Go HTTP Router)

PostgreSQL(database)

database/sql(database access)

bcrypt(Authentication)

<mark>Template Sytem</mark>

Html/template(Go templates)

<mark>Infrastructure Stack</mark>

Nginx(webserver)

