package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type BlogServer struct {
	s *BlogStore
}

func (b *BlogServer) RenderWithBase(content template.HTML) (string, error) {
	ret := &strings.Builder{}
	templ, err := b.s.GetBaseTemplate()
	if err != nil {
		return "", err
	}
	err = templ.Execute(ret, struct {
		Title       string
		PageContent template.HTML
		StyleSheet  template.CSS
	}{
		"BlogSoft", content, b.s.GetStyle(),
	})

	if err != nil {
		return "", err
	}

	return ret.String(), nil
}

func (b *BlogServer) RenderIndex() (string, error) {
	content, err := b.s.GetIndexPage()
	if err != nil {
		return "", fmt.Errorf("blog not properly set up: %w", err)
	}

	page_content := template.HTML(content)
	ret, err := b.RenderWithBase(page_content)
	if err != nil {
		return "", nil
	}
	return ret, nil
}

func (b *BlogServer) RenderBlogPost(name string) (string, error) {
	blog_post, err := b.s.GetBlog(name)
	if err != nil {
		return "", err
	}
	blog_templ, err := b.s.GetBlogPostTemplate()
	if err != nil {
		return "", err
	}

	post_buf := &strings.Builder{}
	date_string := RenderDateString(blog_post.Date)
	err = blog_templ.Execute(post_buf, struct {
		Title   string
		Date    string
		Content template.HTML
	}{
		blog_post.Title,
		date_string,
		template.HTML(blog_post.Content),
	})
	if err != nil {
		return "", err
	}

	page_content := template.HTML(post_buf.String())
	ret, err := b.RenderWithBase(page_content)
	if err != nil {
		return "", nil
	}
	return ret, nil
}

func (b *BlogServer) RenderBlogList() (string, error) {
	posts, err := b.s.GetAllBlogs()
	if err != nil {
		return "", err
	}
	templ, err := b.s.GetBlogListTemplate()
	if err != nil {
		return "", err
	}

	post_buf := &strings.Builder{}
	err = templ.Execute(post_buf, struct {
		Posts []BlogPost
	}{
		posts,
	})
	page_content := template.HTML(post_buf.String())
	ret, err := b.RenderWithBase(page_content)
	if err != nil {
		return "", nil
	}
	return ret, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("port must be provided")
		os.Exit(1)
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	port_str := ":" + strconv.Itoa(port)

	mux := http.NewServeMux()

	serv := BlogServer{&BlogStore{}}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index/", http.StatusSeeOther)
	})

	mux.HandleFunc("/index/", func(w http.ResponseWriter, r *http.Request) {
		index, err := serv.RenderIndex()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write([]byte(index))
	})

	mux.HandleFunc("/blog/", func(w http.ResponseWriter, r *http.Request) {
		target_post := strings.TrimPrefix(r.URL.Path, "/blog/")

		if target_post == "" {
			ret, err := serv.RenderBlogList()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
			}
			_, _ = w.Write([]byte(ret))
			return
		}

		post, err := serv.RenderBlogPost(target_post)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write([]byte(post))
	})

	mux.HandleFunc("/about/", func(w http.ResponseWriter, r *http.Request) {
		ret, err := serv.s.GetAbout()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		ret, err = serv.RenderWithBase(template.HTML(ret))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
		_, _ = w.Write([]byte(ret))
	})

	srv := &http.Server{
		Addr:              port_str,
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

func RenderDateString(date time.Time) string {
	date_string := fmt.Sprintf(
		"%d/%d/%d",
		date.Month(), date.Day(), date.Year(),
	)
	return date_string
}
