package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func faqHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<h1>FAQ Page</h1>
	<ul>
		<li>
			<b>What's this website about?</b>
			It's for Photo management and sharing!
		</li>
		<li>
			<b>How to use this application?</b>
			Sign up, log in, upload and share your photo!
		</li>
	</ul>`)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<h1>Contact Page</h1>
<p>To get in touch, email me at <a href="mailto:c0de-think@protonmail.com">c0de-think@protonmail.com</a></p>`)
}

// r is the info send from browser to our server, which includes: URL, headers, and a body.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>Welcome to Wentao's photo sharing website!</h1>")
}

func main() {
	r := chi.NewRouter()
	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page Not Found", http.StatusNotFound)
	})
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
