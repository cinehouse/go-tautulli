package tautulli

import (
	"context"
	"net/http"
)

// NotificationsService handles communication with the notification related
// methods of the Tautulli API.
type NotificationsService service

// NotifyParameters are parameters for sending a notification using the Tautulli API.
type NotifyParameters struct {
	NotifierID int    `url:"notifier_id"`           // The ID number of the notification agent
	Subject    string `url:"subject"`               // The subject of the message
	Body       string `url:"body"`                  // The body of the message
	Headers    string `url:"headers,omitempty"`     // Optional. The JSON headers for webhook notifications
	ScriptArgs string `url:"script_args,omitempty"` // Optional. The arguments for script notifications
}

const (
	commandNotify = "notify"
)

// Notify sends a notification using the Tautulli API.
func (s *NotificationsService) Notify(ctx context.Context, params *NotifyParameters) (*http.Response, error) {
	req, err := s.client.NewCommand(commandNotify, params)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req)
}
