package app

import (
	"github.com/masterhung0112/go_server/model"
)

func (s *Server) License() *model.License {
	license, _ := s.licenseValue.Load().(*model.License)
	return license
}
