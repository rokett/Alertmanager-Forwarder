# Alertmanager Forwarder
Alertmanager has built-in integrations for a number of services, but if you need to do more then you're looking at using the Webhook functionality.

This forwarder provides a mechanism for forwarding the alert payload from Alertmanager onto either Graylog or vRealize Orchestrator.  Other integrations could easily be added later as required.

## Graylog configuration
You'll need a suitable UDP input to accept the incoming messages.

Each alert is sent as a separate message with the fields being split out for an easier time at the Graylog end.  The following fields are included in the message payload to Graylog for every event.

 - status
 - startsAt
 - endsAt
 - generatorURL
 - labels
 - annotations

## vRO configuration
Each alert is sent on to vRO as a separate POST request.  The workflow called is based on the workflow ID embedded into the ``vro_action`` field of the alert from Alertmanager.  The alert payload is base64 encoded in order to ensure there are no odd issues with a JSON string being embedded into a JSON string; you'll need to decode the base64 string in your workflow.  Your workflow just needs one input called ``json`` of type ``string``.

## Usage
You don't have to use both Graylog and vRO config options.  If you only include one of them, the forwarder will only function for that one.

| Flag            | Description                                                                         | Default Value |
| --------------- | ------------------------------------------------------------------------------------| ------------- |
| graylog-url     | Graylog listener URL; minus the port                                                | none          |
| graylog-port    | Port that the Graylog UDP input is listening on                                     | 0             |
| vro-host        | vRO listener URL; minus the port                                                    | none          |
| vro-port        | Port that the vRO API is listening on                                               | 0             |
| vro-auth        | base64 encoded authorisation header string.  For example 'dmNvdXNlcjpteXBhc3N3b3Jk' | none          |
| port            | Port that the Alertmanager-Forwarder service listens on                             | 10001         |
| version         | Displays application version information and then quits                             | false         |
| help            | Dsplay usage help                                                                   | false         |
| debug           | Enable debugging                                                                    | false         |
| service         | Manage Windows servers.  Possible options are install, uninstall, start, stop       | none          |

Example command:

````
Alertmanager-Forwarder.exe --service install --graylog-url graylog.mydomain.com --graylog-port 12202 --vro-host vro.mydomain.com --vro-port 8281 --vro-auth dmNvYXBpOmFjcVNaa0V6tGE3bGxQUlRwdDVD
````

This will install the service (not started) using the arguments supplied; it will run on the default port of `10001` as the ``port`` flag was not used..  It will be capable of forwarding alerts to Graylog and vRO.  Omit the ``--service`` flag to run it interactively.

## Downloading a release
<https://github.com/rokett/Alertmanager-Forwarder/releases>

## Building the executable
All dependencies are version controlled, so building the project is really easy.

1. ``go get github.com/rokett/alertmanager-forwarder``.
2. From within the repository directory run ``make build``.
3. Hey presto, you have an executable.
