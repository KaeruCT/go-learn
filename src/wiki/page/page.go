package page

import (
    "io/ioutil"
    "os"
)

type Page struct {
    Title string
    Body  []byte
}

func Save(location string, p *Page) error {
    filename := location + p.Title
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func Load(location string, title string) (*Page, error) {
    filename := location + title
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}


func Delete(location string, title string) error {
    filename := location + title
    return os.Remove(filename)
}

func ListAll(location string) ([]*Page, error) {
    fileList, err := ioutil.ReadDir(location)
    pages := make([]*Page, len(fileList))
    if err != nil {
        return nil, err
    }
    for i, fileInfo := range fileList {
        pages[i], _ = Load(location, fileInfo.Name())
    }
    return pages, nil
}
