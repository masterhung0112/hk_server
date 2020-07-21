package web

import (
  "net/http"
  "reflect"
  "runtime"
  "strings"
)

type Handler struct {
  HandleFunc  func(*Context, http.ResponseWriter, *http.Request)
  HandlerName string
}

func GetHandlerName(h func(*Context, http.ResponseWriter, *http.Request)) string {
  handlerName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
  pos := strings.LastIndex(handlerName, ".")
  if pos != -1 && len(handlerName) > pos {
    handlerName = handlerName[pos+1:]
  }
  return handlerName
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  // Generate new request ID

  // Create new context
  c := &Context{}

  // call the real handler
  h.HandleFunc(c, w, r)
}
