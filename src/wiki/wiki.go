package wiki

import (
    "strings"
    "bytes"
    "os"
    "net/http"
    "html/template"
    "wiki/page"
    "regexp"
    "github.com/knieriem/markdown"
)

var templates = template.Must(template.ParseFiles(
    templateLocation + "header.html",
    templateLocation + "footer.html",
    templateLocation + "main.html",
    templateLocation + "edit.html",
    templateLocation + "view.html"))
var titleValidator = regexp.MustCompile("^[a-zA-Z0-9-_]+$")
const templateLocation = "templates/"
const pageSrcLocation = "page-md/"
const pageLocation = "page/"

type htmlPage struct {
    Title string
    Body  template.HTML
    New   bool
}

func renderPageTemplate(w http.ResponseWriter, tmpl string, p *page.Page) {
    err := templates.ExecuteTemplate(w, tmpl + ".html", htmlPage{
        Title: p.Title,
        Body: template.HTML(p.Body),
        New: p.Body == nil,
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func renderListTemplate(w http.ResponseWriter, tmpl string, pages []*page.Page) {
    htmlPages := make([]*htmlPage, len(pages))
    for i, p := range pages {
        htmlPages[i] = &htmlPage{
            Title: p.Title,
            Body: template.HTML(p.Body),
        }
    }

    err := templates.ExecuteTemplate(w, tmpl + ".html", htmlPages)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
    pages, err := page.ListAll(pageLocation)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    renderListTemplate(w, "main", pages)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := page.Load(pageLocation, title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return

    }
    renderPageTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := page.Load(pageSrcLocation, title)
    if err != nil {
        p = &page.Page{Title: title}
    }
    renderPageTemplate(w, "edit", p)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
    title := r.FormValue("title")

    if titleValidator.MatchString(title) {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    } else {
        http.Redirect(w, r, "/", http.StatusFound)
    }
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")

	// save markdown source
	p := &page.Page{Title: title, Body: []byte(body)}
    err := page.Save(pageSrcLocation, p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // save html
    mdParser := markdown.NewParser(&markdown.Extensions{Smart: true})
    mdBody := bytes.NewBuffer(nil)

    mdParser.Markdown(strings.NewReader(body), markdown.ToHTML(mdBody))
    p = &page.Page{Title: title, Body: []byte(mdBody.String())}
    err = page.Save(pageLocation, p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
    page.Delete(pageLocation, title)
    page.Delete(pageSrcLocation, title)

    http.Redirect(w, r, "/", http.StatusFound)
}

func makePageHandler(fn func(http.ResponseWriter, *http.Request, string), title string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        lenPath := len(title)
        title := r.URL.Path[lenPath:]
        if !titleValidator.MatchString(title) {
            http.NotFound(w, r)
            return
        }
        fn(w, r, title)
    }
}

func Start() {
    os.Mkdir(pageSrcLocation, 0777)
    os.Mkdir(pageLocation, 0777)
    http.HandleFunc("/", mainHandler)
    http.HandleFunc("/new/",  newHandler)
    http.HandleFunc("/view/", makePageHandler(viewHandler, "/view/"))
    http.HandleFunc("/edit/", makePageHandler(editHandler, "/edit/"))
    http.HandleFunc("/save/", makePageHandler(saveHandler, "/save/"))
    http.HandleFunc("/delete/", makePageHandler(deleteHandler, "/delete/"))
    http.ListenAndServe(":8080", nil)
}
