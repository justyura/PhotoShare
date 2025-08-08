## About this project

I started learning **Go** in April 2025 and chose to build a lightweight photo-sharing service as my hands-on playground. Each commit captures a concrete lesson—from idiomatic routing with Chi to salted-hash password storage in PostgreSQL—so the code base doubles as both a mini-app *and* a learning log.

### Key Learning Highlights

- **Chi-based routing** – REST-style endpoints with clean middleware stacks.
- **Server-side templates** – safe, dynamic HTML via `html/template`.
- **Architecture choices** – evaluated flat vs. separation-of-concerns vs. dependency-based layouts, then applied an MVC structure.
- **Tailwind CSS** – utility-first styling for a responsive UI.
- **Secure persistence** – bcrypt-hashed, salted passwords in PostgreSQL.
- **Cookie auth & CSRF defense** – cookie-based login wrapped in CSRF-blocking middleware.

### Tech Stack
Go · Chi · Tailwind CSS · PostgreSQL · Docker (dev)

### Roadmap
- Image upload & cloud storage
- User profiles and pagination
- Deployment pipeline (Docker + Kubernetes)

> The app already runs locally, but its true value is showcasing my ability to pick up a new language, weigh design options, and translate security best practices into code. More features will land as I continue refining my Go skills.
test
