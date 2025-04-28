package main

import (
	"bufio"
	"html/template"
	"os"
	"slices"
	"strings"
	"time"
)

type BlogPost struct {
	Title    string
	Filename string
	Date     time.Time
	DateRepr string
	Content  string
}

// TODO
type Cacher struct {
}

type BlogStore struct {
	// c *Cacher
}

func (s *BlogStore) GetBlog(name string) (BlogPost, error) {
	path := "./content/blog/" + name + ".blog"

	file, err := os.ReadFile(path)
	if err != nil {
		return BlogPost{}, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(file)))

	scanner.Scan()
	title := scanner.Text()
	scanner.Scan()
	date := scanner.Text()
	scanner.Scan()
	var content string
	var text string
	scanner.Scan()
	text = scanner.Text()
	for scanner.Scan() {
		content += text + "\n<br>\n"
		text = scanner.Text()
	}
	if text != "" {
		content += text
	}

	if scanner.Err() != nil {
		return BlogPost{}, err
	}

	t, err := time.Parse("1/2/2006", date)
	if err != nil {
		return BlogPost{}, err
	}
	return BlogPost{title, name, t, RenderDateString(t), content}, nil

}

// TODO properly
func (s *BlogStore) GetAllBlogs() ([]BlogPost, error) {
	var entries []BlogPost
	// filepath.WalkDir("./blog/")
	dir, err := os.ReadDir("./content/blog")
	if err != nil {
		return nil, err
	}
	for _, e := range dir {
		if e.IsDir() {
			continue
		}
		blog, err := s.GetBlog(strings.TrimSuffix(e.Name(), ".blog"))
		if err != nil {
			return nil, err
		}
		entries = append(entries, blog)
	}

	return entries, nil
}

func (s *BlogStore) GetLastNBlogs(no int) ([]BlogPost, error) {
	posts, err := s.GetAllBlogs()
	if err != nil {
		return nil, err
	}
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
	return posts[:l], nil
}

func (s *BlogStore) GetAbout() (string, error) {
	path := "./content/about.html"

	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func (s *BlogStore) GetIndexPage() (template.HTML, error) {
	path := "./content/index.html"

	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	t := template.New("index")
	t, err = t.Parse(string(file))
	if err != nil {
		return "", err
	}

	blogs, err := s.GetLastNBlogs(5)
	if err != nil {
		return "", err
	}

	ret := &strings.Builder{}
	err = t.Execute(ret, struct {
		LastPosts []BlogPost
	}{
		blogs,
	})
	if err != nil {
		return "", err
	}

	return template.HTML(ret.String()), nil
}

func (s *BlogStore) GetBaseTemplate() (*template.Template, error) {
	path := "./content/templates/_base.html"
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	t := template.New("base")
	t, err = t.Parse(string(file))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *BlogStore) GetBlogPostTemplate() (*template.Template, error) {
	path := "./content/templates/blog_post.html"
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	t := template.New("base")
	t, err = t.Parse(string(file))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *BlogStore) GetBlogListTemplate() (*template.Template, error) {
	path := "./content/templates/blog_list.html"
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	t := template.New("base")
	t, err = t.Parse(string(file))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *BlogStore) GetStyle() template.CSS {
	path := "./content/style.css"
	file, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return template.CSS(string(file))
}
