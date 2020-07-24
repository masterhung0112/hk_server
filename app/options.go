package app

type Option func(server *Server) error

type AppOption func(a *App)
type AppOptionCreator func() []AppOption