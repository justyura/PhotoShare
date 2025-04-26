package main

import (
	"fmt"
	"net/http"
)

// type Router struct{}
//
// func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	switch r.URL.Path {
// 	case "/":
// 		homeHandler(w, r)
// 	case "/contact":
// 		contactHandler(w, r)
// 	default:
// 		notFoundHandler(w, r)
// 	}
// }

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

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Page Not Found!")
}

func pathHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		homeHandler(w, r)
	case "/contact":
		contactHandler(w, r)
	case "/faq":
		faqHandler(w, r)
	default:
		notFoundHandler(w, r)
	}
}

func main() {
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", http.HandlerFunc(pathHandler))
}
