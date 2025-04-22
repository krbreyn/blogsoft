package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//var blog_dir = os.Getenv("BLOG_DIR")

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index/", http.StatusSeeOther)
	})

	mux.HandleFunc("/index/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Index page!"))
	})

	mux.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {
		target_post := strings.TrimPrefix(r.URL.Path, "/post/")
		post, err := OpenBlogPost(target_post)
		if err != nil {
			w.Write([]byte(target_post))
			w.Write([]byte(err.Error()))
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(RenderPostPage(post)))
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("starting http server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type BlogPost struct {
	Title   string
	Date    string
	Tags    string
	Content string
}

func OpenBlogPost(filename string) (BlogPost, error) {
	path := "./blog/" + filename + ".sbmd"

	_, err := os.Stat(path)
	if err != nil {
		return BlogPost{}, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return BlogPost{}, err
	}

	s := bufio.NewScanner(strings.NewReader(string(file)))

	s.Scan()
	title := s.Text()
	s.Scan()
	date := s.Text()
	s.Scan()
	tags := s.Text()
	s.Scan()
	var content string
	for s.Scan() {
		content += s.Text() + "\n"
	}

	if s.Err() != nil {
		return BlogPost{}, err
	}

	return BlogPost{title, date, tags, content}, nil
}

func RenderIndexPage() {

}

func RenderPostPage(post BlogPost) string {
	var b strings.Builder

	b.WriteString("Post page!\n")
	b.WriteString(post.Title + "\n" + post.Date + "\n" + post.Tags + "\n")
	b.WriteString("\n" + post.Content)

	return b.String()
}

type Cacher interface {
	Is(name string) bool
	Get(name string) BlogPost
	Put(name string) BlogPost
}
