package tautulli

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestNotificationsService_Notify(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		testMethod(t, request, "GET")
		testFormValues(t, request, values{
			"apikey":      "test",
			"cmd":         "notify",
			"notifier_id": "1",
			"subject":     "test",
			"body":        "test",
			"out_type":    "json",
			"callback":    "pong",
			"debug":       "1",
		})
		fmt.Fprint(writer, `[{"number":1}]`)
	})
	ctx := context.Background()
	params := &NotifyParameters{
		NotifierID: 1,
		Subject:    "test",
		Body:       "test",
	}
	_, err := client.Notifications.Notify(ctx, params)
	if err != nil {
		t.Errorf("Notifications.Notify returned error: %v", err)
	}
}
