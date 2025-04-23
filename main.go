package main

import (
	"bufio"
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//var blog_dir = os.Getenv("BLOG_DIR")

func main() {
	base_t := template.New("base")
	base_t = template.Must(base_t.Parse(BaseTmpl))
	index_t := template.New("index")
	index_t = template.Must(index_t.Parse(IndexTmpl))
	post_t := template.New("post")
	post_t = template.Must(post_t.Parse(PostTmpl))

	templates := Templates{base_t, index_t, post_t}
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
		w.Write([]byte(RenderPostPage(post, templates)))
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
	Tags    []string
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
	var tags []string
	tag_line := s.Text()
	if tag_line != "" {
		str := strings.TrimPrefix(tag_line, "[[tags: ")
		if str != tag_line {
			str := strings.TrimSuffix(str, "]]")
			tags = strings.Fields(str)
		}
	}
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

//go:embed templates/base.tmpl
var BaseTmpl string

//go:embed templates/index.tmpl
var IndexTmpl string

//go:embed templates/post.tmpl
var PostTmpl string

type Templates struct {
	Base, Index, Post *template.Template
}

func RenderPostPage(post BlogPost, ts Templates) string {
	b := &strings.Builder{}
	p := &strings.Builder{}
	ts.Post.Execute(p, struct {
		Title, Date, Content string
		Tags                 []string
	}{
		post.Title,
		post.Date,
		post.Content,
		post.Tags,
	})
	post_content := template.HTML(p.String())
	ts.Base.Execute(b, struct {
		Title       string
		PageContent template.HTML
	}{
		post.Title, post_content,
	})
	return b.String()
}

type Cacher interface {
	Is(name string) bool
	Get(name string) BlogPost
	Put(name string) BlogPost
}
