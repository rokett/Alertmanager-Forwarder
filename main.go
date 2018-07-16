package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aphistic/golf"

	"github.com/kardianos/service"
)

var (
	app               = "Alertmanager-Forwarder"
	version           string
	build             string
	gl                = flag.String("graylog-url", "", "Define the Graylog listener URL; minus the port")
	glPort            = flag.Int("graylog-port", 0, "Graylog listener port")
	vroHost           = flag.String("vro-host", "", "Define the vRO host used to build the API URL (including the port).  For example 'vco.mydomain.com:8281'")
	vroAuth           = flag.String("vro-auth", "", "Define the base64 encoded authorisation header string.  For example 'dmNvdXNlcjpteXBhc3N3b3Jk'.")
	vroPort            = flag.Int("vro-port", 0, "vRO listener port")
	svrPort           = flag.Int("port", 10001, "Web server port")
	help              = flag.Bool("help", false, "Display help")
	versionFlg        = flag.Bool("version", false, "Display application version")
	debug             = flag.Bool("debug", false, "Enable debugging?")
	winServiceCommand = flag.String("service", "", "Manage Windows services: install, uninstall, start, stop")
	serviceDesc       = "Forward alerts from Alertmanager to another service"
	logger            service.Logger
)

type program struct{}

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *versionFlg {
		fmt.Printf("%s v%s build %s", app, version, build)
		os.Exit(0)
	}

	if (*winServiceCommand == "install") && (*gl == "" && *vroHost == "") {
		fmt.Println("A forwarding destination MUST be set")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *gl != "" && *glPort == 0 {
		fmt.Println("You MUST set the --graylog-port flag")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *vroHost != "" && *vroPort == 0 {
		fmt.Println("You MUST set the --vro-port flag")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *vroHost != "" && *vroAuth == "" {
		fmt.Println("You MUST set the --vro-auth flag")
		flag.PrintDefaults()
		os.Exit(1)
	}

	svcConfig := &service.Config{
		Name:        app,
		DisplayName: app,
		Description: serviceDesc,
	}

	prg := &program{}

	svc, err := service.New(prg, svcConfig)
	if err != nil {
		log.Println("Error creating new service")
		log.Println(err)

		os.Exit(1)
	}

	errs := make(chan error, 5)
	logger, err = svc.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if *winServiceCommand != "" {
		if *gl != "" {
			(*svcConfig).Arguments = append((*svcConfig).Arguments, "--graylog-url", *gl)
			(*svcConfig).Arguments = append((*svcConfig).Arguments, "--graylog-port", strconv.Itoa(*glPort))
		}

		if *vroHost != "" {
			(*svcConfig).Arguments = append((*svcConfig).Arguments, "--vro-host", *vroHost)
			(*svcConfig).Arguments = append((*svcConfig).Arguments, "--vro-auth", *vroAuth)
		}

		err := service.Control(svc, *winServiceCommand)
		if err != nil {
			msg := fmt.Sprintf("Error creating new service.\n\n%s", err)
			logger.Error(msg)
		}

		os.Exit(0)
	} else {
		err = svc.Run()
		if err != nil {
			msg := fmt.Sprintf("Error running application interactively.\n\n%s", err)
			logger.Error(msg)
		}
	}
}

func (p *program) Start(s service.Service) error {
	// Start should not block; do the actual work async.
	go p.run(s)
	return nil
}

func (p *program) run(s service.Service) {
	if *gl != "" {
		c, err := golf.NewClient()
		if err != nil {
			msg := fmt.Sprintf("Unable to create new golf client.\n\n%s", err)
			logger.Error(msg)

			p.Stop(s)
		}
		defer c.Close()

		url := fmt.Sprintf("udp://%s:%d", *gl, *glPort)
		err = c.Dial(url)
		if err != nil {
			msg := fmt.Sprintf("Unable to dial Graylog input.\n\n%s", err)
			logger.Error(msg)

			p.Stop(s)
		}

		l, err := c.NewLogger()
		if err != nil {
			msg := fmt.Sprintf("Unable to create new golf logger.\n\n%s", err)
			logger.Error(msg)

			p.Stop(s)
		}

		l.SetAttr("app", app)
		l.SetAttr("app_version", version)
		l.SetAttr("app_build", build)

		http.Handle("/graylog", processGL(logger, l))
	}

	if *vroHost != "" {
		http.Handle("/vro", processVRO(logger, *vroHost, *vroPort, *vroAuth))
	}

	listeningPort := fmt.Sprintf(":%d", *svrPort)

	msg := fmt.Sprintf("Starting %s.  Version %s.  Build %s.\n\nListening on port %s", app, version, build, listeningPort)
	logger.Info(msg)

	err := http.ListenAndServe(listeningPort, nil)
	if err != nil {
		logger.Error(err)

		p.Stop(s)
	}
}

func (p *program) Stop(s service.Service) error {
	logger.Info("Stopping")

	return nil
}
