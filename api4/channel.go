package api1

import (
	"net/http"

	"github.com/masterhung0112/hk_server/model"
)

func (api *API) InitChannel() {
	// api.BaseRoutes.Channels.Handle("", api.ApiSessionRequired(getAllChannels)).Methods("GET")
	api.BaseRoutes.Channels.Handle("", api.ApiSessionRequired(createChannel)).Methods("POST")
}

func createChannel(c *Context, w http.ResponseWriter, r *http.Request) {
	channel := model.ChannelFromJson(r.Body)
	if channel == nil {
		c.SetInvalidParam("channel")
		return
	}

	//TODO: Open
	// auditRec := c.MakeAuditRecord("createChannel", audit.Fail)
	// defer c.LogAuditRec(auditRec)
	// auditRec.AddMeta("channel", channel)

	if channel.Type == model.CHANNEL_OPEN && !c.App.SessionHasPermissionToTeam(*c.App.Session(), channel.TeamId, model.PERMISSION_CREATE_PUBLIC_CHANNEL) {
		c.SetPermissionError(model.PERMISSION_CREATE_PUBLIC_CHANNEL)
		return
	}

	if channel.Type == model.CHANNEL_PRIVATE && !c.App.SessionHasPermissionToTeam(*c.App.Session(), channel.TeamId, model.PERMISSION_CREATE_PRIVATE_CHANNEL) {
		c.SetPermissionError(model.PERMISSION_CREATE_PRIVATE_CHANNEL)
		return
	}

	sc, err := c.App.CreateChannelWithUser(channel, c.App.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	//TODO: Open
	// auditRec.Success()
	// auditRec.AddMeta("channel", sc) // overwrite meta
	// c.LogAudit("name=" + channel.Name)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(sc.ToJson()))
}
