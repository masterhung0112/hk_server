package app

func (a *App) UpdateMobileAppBadge(userId string) {
	// select {
	// case a.Srv().PushNotificationsHub.notificationsChan <- PushNotification{
	// 	notificationType: notificationTypeUpdateBadge,
	// 	userId:           userId,
	// }:
	// case <-a.Srv().PushNotificationsHub.stopChan:
	// 	return
	// }
}

func (s *Server) createPushNotificationsHub() {
	// buffer := *s.Config().EmailSettings.PushNotificationBuffer
	// // XXX: This can be _almost_ removed except that there is a dependency with
	// // a.ClearSessionCacheForUser(session.UserId) which invalidates caches,
	// // which then takes to web_hub code. It's a bit complicated, so leaving as is for now.
	// fakeApp := New(ServerConnector(s))
	// hub := PushNotificationsHub{
	// 	notificationsChan: make(chan PushNotification, buffer),
	// 	app:               fakeApp,
	// 	wg:                new(sync.WaitGroup),
	// 	semaWg:            new(sync.WaitGroup),
	// 	sema:              make(chan struct{}, runtime.NumCPU()*8), // numCPU * 8 is a good amount of concurrency.
	// 	stopChan:          make(chan struct{}),
	// 	buffer:            buffer,
	// }
	// go hub.start()
	// s.PushNotificationsHub = hub
}

func (a *App) clearPushNotification(currentSessionId, userId, channelId string) {
	// select {
	// case a.Srv().PushNotificationsHub.notificationsChan <- PushNotification{
	// 	notificationType: notificationTypeClear,
	// 	currentSessionId: currentSessionId,
	// 	userId:           userId,
	// 	channelId:        channelId,
	// }:
	// case <-a.Srv().PushNotificationsHub.stopChan:
	// 	return
	// }
}
