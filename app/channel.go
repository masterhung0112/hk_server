package app

import (
	"net/http"
	"github.com/pkg/errors"
	"github.com/masterhung0112/go_server/store"
	"github.com/masterhung0112/go_server/model"
)


func (a *App) GetChannel(channelId string) (*model.Channel, *model.AppError) {
	channel, err := a.Srv().Store.Channel().Get(channelId, true)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannel", "app.channel.get.existing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannel", "app.channel.get.find.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return channel, nil
}