package wiki

import (
    "strings"
    "bytes"
    "net/http"
    "html/template"
    "wiki/page"
    "regexp"
    "github.com/knieriem/markdown"
)

const lenPath = len("/view/")
var templates = template.Must(template.ParseFiles(
    templateLocation + "edit.html",
    templateLocation + "view.html"))
var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")
const templateLocation = "templates/"
const pageSrcLocation = "page-md/"
const pageLocation = "page/" 

type htmlPage struct {
    Title string
    Body  template.HTML
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *page.Page) {
    err := templates.ExecuteTemplate(w, tmpl + ".html", htmlPage{
        Title: p.Title,
        Body: template.HTML(p.Body),
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := page.Load(pageLocation, title)
	
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
        
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := page.Load(pageSrcLocation, title)
    if err != nil {
        p = &page.Page{Title: title}
    }
    renderTemplate(w, "edit", p)
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

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        title := r.URL.Path[lenPath:]
        if !titleValidator.MatchString(title) {
            http.NotFound(w, r)
            return
        }
        fn(w, r, title)
    }
}

func Start() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.ListenAndServe(":8080", nil)
}
