package app

import "context"

type App struct {
	context context.Context
}

func (a *App) SetContext(c context.Context) {
	a.context = c
}
