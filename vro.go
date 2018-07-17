package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kardianos/service"
)

type vRO struct {
	Parameters []params `json:"parameters"`
}

type params struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Scope string `json:"scope"`
	Value val    `json:"value"`
}

type val struct {
	String str `json:"string"`
}

type str struct {
	Value string `json:"value"`
}

func processVRO(logger service.Logger, host string, port int, auth string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			defer r.Body.Close()

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
				var workflowID string

				for k, v := range alert.Labels.(map[string]interface{}) {
					if k == "vro_action" {
						workflowID = v.(string)
						break
					}
				}

				if workflowID == "" {
					msg := fmt.Sprintf("Alert is missing vro_action; skipping.\n\n%s", alert.Labels)
					logger.Error(msg)

					continue
				}

				alertPL, err := json.Marshal(alert)
				if err != nil {
					msg := fmt.Sprintf("Error marshalling alert struct to payload.\n\n%s", err)
					logger.Error(msg)

					response.Msg = "error marshalling alert struct to payload"
					response.Error = err
					response.Send(http.StatusInternalServerError, w)

					return
				}

				encoded := base64.StdEncoding.EncodeToString(alertPL)

				// Yes the JSON structure for vRO API requests is pretty horrid!
				// str -> val -> params -> full payload
				a := str{
					Value: encoded,
				}

				vroVal := val{
					String: a,
				}

				var vroParams []params
				vroParams = append(vroParams, params{
					Type:  "string",
					Name:  "json",
					Scope: "local",
					Value: vroVal,
				})

				vroPL := vRO{
					Parameters: vroParams,
				}

				payload, err := json.Marshal(vroPL)
				if err != nil {
					msg := fmt.Sprintf("Error marshalling vRO struct to payload.\n\n%s", err)
					logger.Error(msg)

					response.Msg = "error marshalling vRO struct to payload"
					response.Error = err
					response.Send(http.StatusInternalServerError, w)

					return
				}

				url := fmt.Sprintf("https://%s:%d/vco/api/workflows/%s/executions/", host, port, workflowID)

				req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
				if err != nil {
					msg := fmt.Sprintf("Error creating vRO HTTP request object.\n\nURL: %s\n\n%s", url, err)
					logger.Error(msg)

					response.Msg = "error creating vRO HTTP request object"
					response.Error = err
					response.Send(http.StatusInternalServerError, w)

					return
				}

				authHdr := fmt.Sprintf("Basic %s", auth)
				req.Header.Set("Authorization", authHdr)

				req.Header.Set("Content-Type", "application/json")

				msg := fmt.Sprintf("vRO URL: %s\n\nvRO Authorisation Header: %s", url, authHdr)
				logger.Info(msg)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					msg := fmt.Sprintf("Error POSTing to vRO.\n\n%s", err)
					logger.Error(msg)

					response.Msg = "error POSTing to vRO"
					response.Error = err
					response.Send(http.StatusInternalServerError, w)

					return
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)

				switch resp.StatusCode {
				case 201, 202:
					logger.Info("POST accepted by vRO.\n\nWorkflow execution started.")
				case 400:
					logger.Error("Unknown error from vRO.\n\nResponse\n" + string(body))

					response.Msg = fmt.Sprintf("Unknown error from vRO.  %s", string(body))
					response.Send(http.StatusInternalServerError, w)

					return
				case 401:
					logger.Error("User not authorised to access vRO API")

					response.Msg = "User not authorised to access vRO API"
					response.Send(http.StatusInternalServerError, w)

					return
				default:
					msg := fmt.Sprintf("Unknown status code from vRO.\n\nStatus Code: %d", resp.StatusCode)
					logger.Info(msg)

					return
				}
			}
		default:
			msg := fmt.Sprintf("Invalid HTTP method called; %s", r.Method)
			logger.Error(msg)

			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
}
