package pkg

import (
	"bytes"
	"html/template"
)

const (
	MARKDOWN string = "MarkdownV2"
	HTML     string = "HTML"
)

func Parse(fileName string, data interface{}) (string, error) {
	t, err := template.ParseFiles(fileName)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
