package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

//go:embed templates/base.tmpl
var BaseTmpl string

//go:embed templates/post.tmpl
var PostTmpl string

type Templates struct {
	Base, Post *template.Template
}

type BlogStore interface {
	Get(fillename string) (BlogPost, error)
	GetAll() ([]BlogPost, error)
	GetLastN(n int) ([]BlogPost, error)
}

type Cacher interface {
	Is(name string) bool
	Get(name string) BlogPost
	Put(name string) BlogPost
}

type BlogServer struct {
	s BlogStore
	c Cacher
}

//var blog_dir = os.Getenv("BLOG_DIR")

func main() {
	base_t := template.New("base")
	base_t = template.Must(base_t.Parse(BaseTmpl))
	post_t := template.New("post")
	post_t = template.Must(post_t.Parse(PostTmpl))

	templates := Templates{base_t, post_t}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index/", http.StatusSeeOther)
	})

	mux.HandleFunc("/index/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(RenderIndexPage(templates)))
	})

	mux.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {
		target_post := strings.TrimPrefix(r.URL.Path, "/post/")
		post, err := OpenBlogPost(target_post)
		if err != nil {
			_, _ = w.Write([]byte(target_post))
			_, _ = w.Write([]byte(err.Error()))
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(RenderPostPage(post, templates)))
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
	Title    string
	Filename string
	Date     time.Time
	Tags     []string
	Content  string
}

func OpenBlogPost(filename string) (BlogPost, error) {
	path := "./blog/" + filename + ".sbml"
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
	var text string
	s.Scan()
	text = s.Text()
	for s.Scan() {
		content += text + "\n<br>\n"
		text = s.Text()
	}
	if text != "" {
		content += text
	}

	if s.Err() != nil {
		return BlogPost{}, err
	}

	t, err := time.Parse("1/2/2006", date)
	if err != nil {
		return BlogPost{}, err
	}
	return BlogPost{title, filename, t, tags, content}, nil
}

func OpenIndex() (string, error) {
	path := "./blog/index.sbml"
	_, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	s := bufio.NewScanner(strings.NewReader(string(file)))
	var ret string
	for s.Scan() {
		text := s.Text()
		if strings.HasPrefix(text, "[[blog_last_x") && strings.HasSuffix(text, "]]") {
			split := strings.Fields(text)
			if len(split) == 2 {
				no, err := strconv.Atoi(strings.TrimSuffix(strings.Fields(text)[1], "]]"))
				if err == nil {
					text = BlogLastX(no)
				}
			}
		}
		ret += text + "\n"
	}
	return ret, nil
}

func GetStyle() template.CSS {
	path := "./blog/style.css"
	_, err := os.Stat(path)
	if err != nil {
		return ""
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return template.CSS(string(file))
}

// TODO properly
func GetAllPosts() []BlogPost {
	// var entries []string
	// filepath.WalkDir("./blog/")
	b1, _ := OpenBlogPost("test_post")
	b2, _ := OpenBlogPost("test_2")
	b3, _ := OpenBlogPost("test_3")
	return []BlogPost{b1, b2, b3}
}

func BlogLastX(no int) string {
	// lazy, unoptimized
	// TODO open each file one by one and just check the date

	posts := GetAllPosts()
	slices.SortFunc(posts, func(i, j BlogPost) int {
		if i.Date.After(j.Date) {
			return -1
		}
		if i.Date.Before(j.Date) {
			return 1
		}
		return 0
	})
	l := min(len(posts), no)
	posts = posts[:l]
	var ret string
	for _, p := range posts {
		ret += fmt.Sprintf(`<a href="/post/%s">%s</a><br>`, p.Filename, p.Title)
		ret += "&emsp;"
		ret += RenderDateString(p.Date)
		ret += "<br>\n"
	}
	return ret
}

func RenderPostPage(post BlogPost, ts Templates) string {
	b := &strings.Builder{}
	p := &strings.Builder{}
	date_string := RenderDateString(post.Date)
	_ = ts.Post.Execute(p, struct {
		Title   string
		Date    string
		Content template.HTML
		Tags    []string
	}{
		post.Title,
		date_string,
		template.HTML(post.Content),
		post.Tags,
	})
	page_content := template.HTML(p.String())
	_ = ts.Base.Execute(b, struct {
		Title       string
		PageContent template.HTML
		StyleSheet  template.CSS
	}{
		post.Title, page_content, GetStyle(),
	})
	return b.String()
}

func RenderIndexPage(ts Templates) string {
	b := &strings.Builder{}
	index, err := OpenIndex()
	// hack?
	// TODO CHECK TEMPLATES FOR ERRORS
	if err == nil {
		_ = ts.Base.Execute(b, struct {
			Title       string
			PageContent template.HTML
			StyleSheet  template.CSS
		}{
			"BlogSoft", template.HTML(index), GetStyle(),
		})
	} else {
		_ = ts.Base.Execute(b, struct {
			Title       string
			PageContent template.HTML
			StyleSheet  template.CSS
		}{
			"BlogSoft", template.HTML("no index page found"), GetStyle(),
		})
	}

	return b.String()
}

func RenderDateString(date time.Time) string {
	date_string := fmt.Sprintf(
		"%d/%d/%d",
		date.Month(), date.Day(), date.Year(),
	)
	return date_string
}
