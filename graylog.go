package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kardianos/service"

	"github.com/aphistic/golf"
)

func processGL(logger service.Logger, l *golf.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var response response

			var alertGroup alertGroup

			err := json.NewDecoder(r.Body).Decode(&alertGroup)
			if err != nil {
				msg := fmt.Sprintf("Error decoding json body from Alertmanager.\n\n%s", err)
				logger.Error(msg)

				response.Msg = "error decoding json body from AlertManager"
				response.Error = err
				response.Send(http.StatusBadRequest, w)

				return
			}

			for _, alert := range alertGroup.Alerts {
				payload := map[string]interface{}{
					"status":       alert.Status,
					"startsAt":     alert.StartsAt,
					"endsAt":       alert.EndsAt,
					"generatorURL": alert.GeneratorURL,
				}

				for k, v := range alert.Labels.(map[string]interface{}) {
					payload["label_"+k] = v
				}

				for k, v := range alert.Annotations.(map[string]interface{}) {
					payload["annotation_"+k] = v
				}

				l.Infom(payload, "Alert from Alertmanager")
			}
		default:
			msg := fmt.Sprintf("Invalid HTTP method called; %s", r.Method)
			logger.Error(msg)

			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
}
