package utils

import (
	"net/url"
	"html/template"
	"fmt"
	"crypto/rand"
	"encoding/base64"
	"path"
	"crypto"
	"net/http"
	"github.com/masterhung0112/go_server/model"
)

func RenderWebAppError(config *model.Config, w http.ResponseWriter, r *http.Request, err *model.AppError, s crypto.Signer) {
	RenderWebError(config, w, r, err.StatusCode, url.Values{
		"message": []string{err.Message},
	}, s)
}

func RenderWebError(config *model.Config, w http.ResponseWriter, r *http.Request, status int, params url.Values, s crypto.Signer) {
	queryString := params.Encode()

	subpath, _ := GetSubpathFromConfig(config)

	h := crypto.SHA256
	sum := h.New()
	sum.Write([]byte(path.Join(subpath, "error") + "?" + queryString))
	signature, err := s.Sign(rand.Reader, sum.Sum(nil), h)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	destination := path.Join(subpath, "error") + "?" + queryString + "&s=" + base64.URLEncoding.EncodeToString(signature)

	if status >= 300 && status < 400 {
		http.Redirect(w, r, destination, status)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	fmt.Fprintln(w, `<!DOCTYPE html><html><head></head>`)
	fmt.Fprintln(w, `<body onload="window.location = '`+template.HTMLEscapeString(template.JSEscapeString(destination))+`'">`)
	fmt.Fprintln(w, `<noscript><meta http-equiv="refresh" content="0; url=`+template.HTMLEscapeString(destination)+`"></noscript>`)
	fmt.Fprintln(w, `<a href="`+template.HTMLEscapeString(destination)+`" style="color: #c0c0c0;">...</a>`)
	fmt.Fprintln(w, `</body></html>`)
}