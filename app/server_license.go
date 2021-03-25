package app

import (
	"github.com/masterhung0112/hk_server/v5/model"
)

func (s *Server) License() *model.License {
	license, _ := s.licenseValue.Load().(*model.License)
	return license
}
