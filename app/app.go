package app

import (
	"context"
	"github.com/masterhung0112/hk_server/einterfaces"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/services/httpservice"
	"github.com/masterhung0112/hk_server/services/mailservice"
	"github.com/masterhung0112/hk_server/services/searchengine"
	"github.com/masterhung0112/hk_server/utils"
	"github.com/mattermost/go-i18n/i18n"
	"html/template"
	"net/http"
	"strconv"

	"github.com/masterhung0112/hk_server/model"
	goi18n "github.com/mattermost/go-i18n/i18n"
)

type App struct {
	srv *Server

	// XXX: This is required because removing this needs BleveEngine
	// to be registered in (h *MainHelper) setupStore, but that creates
	// a cyclic dependency as bleve tests themselves import testlib.
	searchEngine *searchengine.Broker

	context context.Context

	t goi18n.TranslateFunc

	session        model.Session
	requestId      string
	ipAddress      string
	path           string
	userAgent      string
	acceptLanguage string
}

func New(options ...AppOption) *App {
	app := &App{}

	for _, option := range options {
		option(app)
	}

	return app
}

func (a *App) SetContext(c context.Context) {
	a.context = c
}

func (a *App) RequestLicenseAndAckWarnMetric(warnMetricId string, isBot bool) *model.AppError {
	if *a.Config().ExperimentalSettings.RestrictSystemAdmin {
		return model.NewAppError("RequestLicenseAndAckWarnMetric", "api.restricted_system_admin", nil, "", http.StatusForbidden)
	}

	currentUser, appErr := a.GetUser(a.Session().UserId)
	if appErr != nil {
		return appErr
	}

	registeredUsersCount, err := a.Srv().Store.User().Count(model.UserCountOptions{})
	if err != nil {
		mlog.Error("Error retrieving the number of registered users", mlog.Err(err))
		return model.NewAppError("RequestLicenseAndAckWarnMetric", "api.license.request_trial_license.fail_get_user_count.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	trialLicenseRequest := &model.TrialLicenseRequest{
		ServerID:              a.TelemetryId(),
		Name:                  currentUser.GetDisplayName(model.SHOW_FULLNAME),
		Email:                 currentUser.Email,
		SiteName:              *a.Config().TeamSettings.SiteName,
		SiteURL:               *a.Config().ServiceSettings.SiteURL,
		Users:                 int(registeredUsersCount),
		TermsAccepted:         true,
		ReceiveEmailsAccepted: true,
	}

	if trialLicenseRequest.SiteURL == "" {
		return model.NewAppError("RequestLicenseAndAckWarnMetric", "api.license.request_trial_license.no-site-url.app_error", nil, "", http.StatusBadRequest)
	}

	if err := a.Srv().RequestTrialLicense(trialLicenseRequest); err != nil {
		// turn off warn metric warning even in case of StartTrial failure
		if nerr := a.setWarnMetricsStatusAndNotify(warnMetricId); nerr != nil {
			return nerr
		}

		return err
	}

	if appErr = a.NotifyAndSetWarnMetricAck(warnMetricId, currentUser, true, isBot); appErr != nil {
		return appErr
	}

	return nil
}

func (a *App) NotifyAndSetWarnMetricAck(warnMetricId string, sender *model.User, forceAck bool, isBot bool) *model.AppError {
	if warnMetric, ok := model.WarnMetricsTable[warnMetricId]; ok {
		data, nErr := a.Srv().Store.System().GetByName(warnMetric.Id)
		if nErr == nil && data != nil && data.Value == model.WARN_METRIC_STATUS_ACK {
			mlog.Debug("This metric warning has already been acknowledged", mlog.String("id", warnMetric.Id))
			return nil
		}

		if !forceAck {
			if len(*a.Config().EmailSettings.SMTPServer) == 0 {
				return model.NewAppError("NotifyAndSetWarnMetricAck", "api.email.send_warn_metric_ack.missing_server.app_error", nil, utils.T("api.context.invalid_param.app_error", map[string]interface{}{"Name": "SMTPServer"}), http.StatusInternalServerError)
			}
			T := utils.GetUserTranslations(sender.Locale)
			bodyPage := a.Srv().EmailService.newEmailTemplate("warn_metric_ack", sender.Locale)
			bodyPage.Props["ContactNameHeader"] = T("api.templates.warn_metric_ack.body.contact_name_header")
			bodyPage.Props["ContactNameValue"] = sender.GetFullName()
			bodyPage.Props["ContactEmailHeader"] = T("api.templates.warn_metric_ack.body.contact_email_header")
			bodyPage.Props["ContactEmailValue"] = sender.Email

			//same definition as the active users count metric displayed in the SystemConsole Analytics section
			registeredUsersCount, cerr := a.Srv().Store.User().Count(model.UserCountOptions{})
			if cerr != nil {
				mlog.Error("Error retrieving the number of registered users", mlog.Err(cerr))
			} else {
				bodyPage.Props["RegisteredUsersHeader"] = T("api.templates.warn_metric_ack.body.registered_users_header")
				bodyPage.Props["RegisteredUsersValue"] = registeredUsersCount
			}
			bodyPage.Props["SiteURLHeader"] = T("api.templates.warn_metric_ack.body.site_url_header")
			bodyPage.Props["SiteURL"] = a.GetSiteURL()
			bodyPage.Props["TelemetryIdHeader"] = T("api.templates.warn_metric_ack.body.diagnostic_id_header")
			bodyPage.Props["TelemetryIdValue"] = a.TelemetryId()
			bodyPage.Props["Footer"] = T("api.templates.warn_metric_ack.footer")

			warnMetricStatus, warnMetricDisplayTexts := a.getWarnMetricStatusAndDisplayTextsForId(warnMetricId, T, false)
			if warnMetricStatus == nil {
				return model.NewAppError("NotifyAndSetWarnMetricAck", "api.email.send_warn_metric_ack.invalid_warn_metric.app_error", nil, "", http.StatusInternalServerError)
			}

			subject := T("api.templates.warn_metric_ack.subject")
			bodyPage.Props["Title"] = warnMetricDisplayTexts.EmailBody

			if err := mailservice.SendMailUsingConfig(model.MM_SUPPORT_ADVISOR_ADDRESS, subject, bodyPage.Render(), a.Config(), false, sender.Email); err != nil {
				mlog.Error("Error while sending email", mlog.String("destination email", model.MM_SUPPORT_ADVISOR_ADDRESS), mlog.Err(err))
				return model.NewAppError("NotifyAndSetWarnMetricAck", "api.email.send_warn_metric_ack.failure.app_error", map[string]interface{}{"Error": err.Error()}, "", http.StatusInternalServerError)
			}
		}

		if err := a.setWarnMetricsStatusAndNotify(warnMetric.Id); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Srv() *Server {
	return a.srv
}
func (a *App) Log() *mlog.Logger {
	return a.srv.Log
}
func (a *App) NotificationsLog() *mlog.Logger {
	return a.srv.NotificationsLog
}
func (a *App) T(translationID string, args ...interface{}) string {
	return a.t(translationID, args...)
}
func (a *App) Ldap() einterfaces.LdapInterface {
	return a.srv.Ldap
}
func (a *App) Metrics() einterfaces.MetricsInterface {
	return a.srv.Metrics
}
func (a *App) Cloud() einterfaces.CloudInterface {
	return a.srv.Cloud
}
func (a *App) HTTPService() httpservice.HTTPService {
	return a.srv.HTTPService
}

func (a *App) TelemetryId() string {
	return a.Srv().TelemetryId()
}

func (s *Server) HTMLTemplates() *template.Template {
	if s.htmlTemplateWatcher != nil {
		return s.htmlTemplateWatcher.Templates()
	}

	return nil
}

func (a *App) Handle404(w http.ResponseWriter, r *http.Request) {
	ipAddress := utils.GetIpAddress(r, a.Config().ServiceSettings.TrustedProxyIPHeader)
	mlog.Debug("not found handler triggered", mlog.String("path", r.URL.Path), mlog.Int("code", 404), mlog.String("ip", ipAddress))

	if *a.Config().ServiceSettings.WebserverMode == "disabled" {
		http.NotFound(w, r)
		return
	}

	utils.RenderWebAppError(a.Config(), w, r, model.NewAppError("Handle404", "api.context.404.app_error", nil, "", http.StatusNotFound), a.AsymmetricSigningKey())
}

func (a *App) InitServer() {
	a.srv.AppInitializedOnce.Do(func() {
		a.DoAppMigrations()
	})
}

func (a *App) SetServer(srv *Server) {
	a.srv = srv
}

func (a *App) Session() *model.Session {
	return &a.session
}

func (a *App) RequestId() string {
	return a.requestId
}
func (a *App) IpAddress() string {
	return a.ipAddress
}
func (a *App) Path() string {
	return a.path
}
func (a *App) UserAgent() string {
	return a.userAgent
}
func (a *App) AcceptLanguage() string {
	return a.acceptLanguage
}
func (a *App) SetSession(s *model.Session) {
	a.session = *s
}
func (a *App) SetT(t goi18n.TranslateFunc) {
	a.t = t
}

func (s *Server) getSystemInstallDate() (int64, *model.AppError) {
	systemData, err := s.Store.System().GetByName(model.SYSTEM_INSTALLATION_DATE_KEY)
	if err != nil {
		return 0, model.NewAppError("getSystemInstallDate", "app.system.get_by_name.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	value, err := strconv.ParseInt(systemData.Value, 10, 64)
	if err != nil {
		return 0, model.NewAppError("getSystemInstallDate", "app.system_install_date.parse_int.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return value, nil
}

func (a *App) Cluster() einterfaces.ClusterInterface {
	return a.srv.Cluster
}
func (a *App) SearchEngine() *searchengine.Broker {
	return a.searchEngine
}

func (a *App) setWarnMetricsStatusAndNotify(warnMetricId string) *model.AppError {
	// Ack all metric warnings on the server
	if err := a.setWarnMetricsStatus(model.WARN_METRIC_STATUS_ACK); err != nil {
		return err
	}

	// Inform client that this metric warning has been acked
	message := model.NewWebSocketEvent(model.WEBSOCKET_WARN_METRIC_STATUS_REMOVED, "", "", "", nil)
	message.Add("warnMetricId", warnMetricId)
	a.Publish(message)

	return nil
}

func (a *App) setWarnMetricsStatus(status string) *model.AppError {
	mlog.Debug("Set monitoring status for all warn metrics", mlog.String("status", status))
	for _, warnMetric := range model.WarnMetricsTable {
		if err := a.setWarnMetricsStatusForId(warnMetric.Id, status); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) setWarnMetricsStatusForId(warnMetricId string, status string) *model.AppError {
	mlog.Debug("Store status for warn metric", mlog.String("warnMetricId", warnMetricId), mlog.String("status", status))
	if err := a.Srv().Store.System().SaveOrUpdateWithWarnMetricHandling(&model.System{
		Name:  warnMetricId,
		Value: status,
	}); err != nil {
		mlog.Error("Unable to write to database.", mlog.Err(err))
		return model.NewAppError("setWarnMetricsStatusForId", "app.system.warn_metric.store.app_error", map[string]interface{}{"WarnMetricName": warnMetricId}, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (a *App) getWarnMetricStatusAndDisplayTextsForId(warnMetricId string, T i18n.TranslateFunc, isE0Edition bool) (*model.WarnMetricStatus, *model.WarnMetricDisplayTexts) {
	var warnMetricStatus *model.WarnMetricStatus
	var warnMetricDisplayTexts = &model.WarnMetricDisplayTexts{}

	if warnMetric, ok := model.WarnMetricsTable[warnMetricId]; ok {
		warnMetricStatus = &model.WarnMetricStatus{
			Id:    warnMetric.Id,
			Limit: warnMetric.Limit,
			Acked: false,
		}

		if T == nil {
			mlog.Debug("No translation function")
			return warnMetricStatus, nil
		}

		warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.bot_response.notification_success.message")

		switch warnMetricId {
		case model.SYSTEM_WARN_METRIC_NUMBER_OF_TEAMS_5:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.number_of_teams_5.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_teams_5.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.number_of_teams_5.start_trial_notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.number_of_teams_5.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_teams_5.notification_body")
			}
		case model.SYSTEM_WARN_METRIC_MFA:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.mfa.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.mfa.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.mfa.start_trial_notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.mfa.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.mfa.notification_body")
			}
		case model.SYSTEM_WARN_METRIC_EMAIL_DOMAIN:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.email_domain.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.email_domain.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.email_domain.start_trial_notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.email_domain.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.email_domain.notification_body")
			}
		case model.SYSTEM_WARN_METRIC_NUMBER_OF_CHANNELS_50:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.number_of_channels_50.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_channels_50.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.number_of_channels_50.start_trial.notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.number_of_channels_50.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_channels_50.notification_body")
			}
		case model.SYSTEM_WARN_METRIC_NUMBER_OF_ACTIVE_USERS_100:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.number_of_active_users_100.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_active_users_100.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.number_of_active_users_100.start_trial.notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.number_of_active_users_100.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_active_users_100.notification_body")
			}
		case model.SYSTEM_WARN_METRIC_NUMBER_OF_ACTIVE_USERS_200:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.number_of_active_users_200.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_active_users_200.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.number_of_active_users_200.start_trial.notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.number_of_active_users_200.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_active_users_200.notification_body")
			}
		case model.SYSTEM_WARN_METRIC_NUMBER_OF_ACTIVE_USERS_300:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.number_of_active_users_300.start_trial.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_active_users_300.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.number_of_active_users_300.start_trial.notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.number_of_active_users_300.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_active_users_300.notification_body")
			}
		case model.SYSTEM_WARN_METRIC_NUMBER_OF_ACTIVE_USERS_500:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.number_of_active_users_500.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_active_users_500.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.number_of_active_users_500.start_trial.notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.number_of_active_users_500.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_active_users_500.notification_body")
			}
		case model.SYSTEM_WARN_METRIC_NUMBER_OF_POSTS_2M:
			warnMetricDisplayTexts.BotTitle = T("api.server.warn_metric.number_of_posts_2M.notification_title")
			if isE0Edition {
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_posts_2M.start_trial.notification_body")
				warnMetricDisplayTexts.BotSuccessMessage = T("api.server.warn_metric.number_of_posts_2M.start_trial.notification_success.message")
			} else {
				warnMetricDisplayTexts.EmailBody = T("api.server.warn_metric.number_of_posts_2M.contact_us.email_body")
				warnMetricDisplayTexts.BotMessageBody = T("api.server.warn_metric.number_of_posts_2M.notification_body")
			}
		default:
			mlog.Error("Invalid metric id", mlog.String("id", warnMetricId))
			return nil, nil
		}

		return warnMetricStatus, warnMetricDisplayTexts
	}
	return nil, nil
}
