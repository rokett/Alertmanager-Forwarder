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

type vroError struct {
	Alert map[string]interface{}
	Msg   string
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

			var vroErrors []vroError

			for _, alert := range alertGroup.Alerts {
				var workflowID string

				al := mapAlert(alert)

				for k, v := range alert.Labels.(map[string]interface{}) {
					if k == "vro_action" {
						workflowID = v.(string)
						break
					}
				}

				if workflowID == "" {
					msg := fmt.Sprintf("Alert is missing vro_action; skipping.\n\n%s", alert.Labels)
					logger.Error(msg)

					vroErrors = append(vroErrors, vroError{
						Alert: al,
						Msg:   "Alert is missing vro_action; skipping",
					})

					continue
				}

				alertPL, err := json.Marshal(alert)
				if err != nil {
					msg := fmt.Sprintf("Error marshalling alert struct to payload.\n\n%s", err)
					logger.Error(msg)

					vroErrors = append(vroErrors, vroError{
						Alert: al,
						Msg:   msg,
					})

					continue
				}

				encoded := base64.StdEncoding.EncodeToString(alertPL)

				// Yes the JSON structure for vRO API requests is pretty horrid!
				// str -> val -> params -> full payload.  Example below with a truncated base64 encoded value field.
				/*
					{
						"parameters": [
							{
								"type": "string",
								"name": "json",
								"scope": "local",
								"value": {
									"string": {
										"value": "eyJzdGF0dXM"
									}
								}
							}
						]
					}
				*/
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

					vroErrors = append(vroErrors, vroError{
						Alert: al,
						Msg:   msg,
					})

					continue
				}

				url := fmt.Sprintf("https://%s:%d/vco/api/workflows/%s/executions/", host, port, workflowID)

				req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
				if err != nil {
					msg := fmt.Sprintf("Error creating vRO HTTP request object.\n\nURL: %s\n\n%s", url, err)
					logger.Error(msg)

					vroErrors = append(vroErrors, vroError{
						Alert: al,
						Msg:   msg,
					})

					continue
				}

				authHdr := fmt.Sprintf("Basic %s", auth)
				req.Header.Set("Authorization", authHdr)

				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					msg := fmt.Sprintf("Error POSTing to vRO.\n\n%s", err)
					logger.Error(msg)

					vroErrors = append(vroErrors, vroError{
						Alert: al,
						Msg:   msg,
					})

					continue
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)

				switch resp.StatusCode {
				case 201, 202:
					logger.Info("POST accepted by vRO.\n\nWorkflow execution started.")
				case 400:
					msg := fmt.Sprintf("Unknown error from vRO.\n\nResponse\n%s", string(body))
					logger.Error(msg)

					vroErrors = append(vroErrors, vroError{
						Alert: al,
						Msg:   msg,
					})

					continue
				case 401:
					msg := fmt.Sprint("User not authorised to access vRO API")
					logger.Error(msg)

					vroErrors = append(vroErrors, vroError{
						Alert: al,
						Msg:   msg,
					})

					continue
				default:
					msg := fmt.Sprintf("Unknown status code from vRO.\n\nStatus Code: %d", resp.StatusCode)
					logger.Info(msg)

					vroErrors = append(vroErrors, vroError{
						Alert: al,
						Msg:   msg,
					})

					continue
				}
			}

			if len(vroErrors) > 0 {
				response.Errors = vroErrors
				response.Send(http.StatusInternalServerError, w)
			}
		default:
			msg := fmt.Sprintf("Invalid HTTP method called; %s", r.Method)
			logger.Error(msg)

			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
}
