package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/apooravm/multi-serve/src/routes"
	"github.com/apooravm/multi-serve/src/subdomains"
	"github.com/apooravm/multi-serve/src/utils"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	// "golang.org/x/crypto/ssh"
)

type (
	Host struct {
		Echo *echo.Echo
	}
)

var (
	ticker     = time.NewTicker(8 * time.Minute)
	quit       = make(chan struct{})
	DEV_MODE   = false
	DOMAIN_URL = "apooravm.xyz"
)

func startHTTPServer() {
	// dev mode
	if len(os.Args) > 1 {
		if os.Args[1] == "dev" {
			DEV_MODE = true
			log.Println("Dev mode")
			if err := godotenv.Load("./secrets/.env"); err != nil {
				log.Println("Error loading .env file")
			}
		}
	}

	// Init vars
	utils.InitGlobalVars()
	utils.InitDirs()
	utils.S3_ObjectInfoArr()
	utils.InitFiles()
	utils.InitSetupFunc()

	// Download files; resume and such
	if err := utils.S3_DownloadFiles(); err != nil {
		utils.LogData("main.go err_id:001 | error downloading S3 files", err.Error())

	} else {
		utils.LogData("S3 Files downloaded successfully 🎉")
	}

	// Download notes data
	if err := utils.DownloadAndWriteNoteData(); err != nil {
		utils.LogData("main.go err_id:002 | error downloading note files", err.Error())

	} else {
		utils.LogData("S3 Notes downloaded successfully 🙌")
	}

	PORT := utils.PORT
	utils.LogData("Live on PORT", PORT, "🔥")

	// main server code
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	// resolve subdomains
	e.Use(subdomains.SubdomainMiddleware())

	// blog routes
	e.Host("blog.*").Any("/*", func(c echo.Context) error {
		return c.String(http.StatusOK, "Blog")
	})

	// API routes
	api := e.Group("")
	api.Use(middleware.CORS())
	api.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		BeforeNextFunc: func(c echo.Context) {
			c.Set("customValueOnRequest", 42)
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			_ = c.Get("customValueOnRequest")
			data := utils.Log{
				ContentLength: v.ContentLength,
				Error:         v.Error,
				Host:          v.Host,
				Latency:       v.Latency,
				RemoteIP:      v.RemoteIP,
				ResponseSize:  v.ResponseSize,
				Time:          v.StartTime,
				Status:        v.Status,
				URI:           v.URI,
				Protocol:      v.Protocol,
			}
			if err := utils.AppendLogToFile(&data, utils.REQUEST_LOG_PATH); err != nil {
				utils.LogData("main.go err_id:003 |", err.Error())
			}
			return nil
		},
	}))

	api.GET("/help", handleNotesRes)
	DefaultGroup(api)

	// fallback?
	e.Static("/", "public")
	e.Any("/*", func(c echo.Context) error {
		sub := c.Get("subdomain")

		log.Println("SUBDOMIAN HERE", sub)

		if sub == nil {
			return c.String(http.StatusOK, "Main app")
		}

		return c.String(http.StatusOK, fmt.Sprintf("Tenant: %s", sub))
	})
	e.Logger.Fatal(e.Start(":" + PORT))
}

// Executed every 5 minutes
// mainly to get around renders idle time limit
func cron_jobs() {
	api_endpoint := "https://" + DOMAIN_URL + "/api/ping"
	if DEV_MODE {
		api_endpoint = fmt.Sprintf("http://localhost:%s", utils.PORT)
	}

	_, err := http.Get(api_endpoint)
	if err != nil {
		fmt.Println("Error pinging. %s", err.Error())
	} else {
		// fmt.Println(res.Status)
	}
}

func idk() {
	fmt.Println("Starting CRON jobs")
	for {
		select {
		case <-ticker.C:
			cron_jobs()

		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// Render only supports http servers
// to make a proper UDP/SSH/TCP/whatever compatible one
// need to host on a compute instance
// wip branch
func main() {
	go func() {
		idk()
	}()

	startHTTPServer()
	return

	// var wg sync.WaitGroup
	// wg.Add(2)
	//
	// go func() {
	// 	defer wg.Done()
	// 	startHTTPServer()
	// }()
	//
	// go func() {
	// 	defer wg.Done()
	// 	startSSHServer()
	// }()
	//
	// log.Println("Servers running...")
	// wg.Wait()
}

func handleNotesRes(c echo.Context) error {
	return c.File("./helper.txt")
}

func DefaultGroup(group *echo.Group) {
	routes.ApiGroup(group.Group("/api"))
}

// func startSSHServer() {
// 	pvt_key_loc := "/etc/secrets/multi_serve_ssh_key"
// 	if len(os.Args) > 1 {
// 		if os.Args[1] == "dev" {
// 			pvt_key_loc = "./secrets/multi_serve_ssh_key"
// 		}
// 	}
// 	fmt.Println("SSH Server Started...")
//
// 	// Reusing the same generated key everytime
// 	// ssh-keygen -lv -f ssh_host_key.pub
// 	pvtBytes, err := os.ReadFile(pvt_key_loc)
// 	if err != nil {
// 		log.Fatal("Failed to load private key:", err.Error())
// 	}
//
// 	privateKey, err := gossh.ParsePrivateKey(pvtBytes)
// 	if err != nil {
// 		log.Fatal("Failed to parse private key:", err.Error())
// 	}
//
// 	// Set up the SSH server with authentication
// 	ssh.Handle(func(s ssh.Session) {
// 		io.WriteString(s, "Welcome to my coolass ssh server.\n")
// 		io.WriteString(s, "Enter smn idk:\n")
//
// 		if _, _, isPty := s.Pty(); isPty {
// 			term := term.NewTerminal(s, "> ")
// 			for {
// 				line, err := term.ReadLine()
// 				if err != nil {
// 					break
// 				}
// 				switch line {
// 				case "exit":
// 					term.Write([]byte("GG\n"))
// 					return
//
// 				default:
// 					term.Write([]byte(fmt.Sprintf("You said: %s\n", line)))
// 				}
// 			}
// 		} else {
// 			scanner := bufio.NewScanner(s)
// 			io.WriteString(s, "Enter smn idk:\n")
//
// 			for scanner.Scan() {
// 				text := scanner.Text()
// 				switch text {
// 				case "exit":
// 					io.WriteString(s, "GG\n")
// 					return
//
// 				default:
// 					io.WriteString(s, "You said: "+text+"\n")
// 				}
// 			}
//
// 			if err := scanner.Err(); err != nil {
// 				log.Println("Error reading ssh input:", err.Error())
// 			}
// 		}
//
// 		// scanner := bufio.NewScanner(s)
// 		// for scanner.Scan() {
// 		// 	text := scanner.Text()
// 		// 	switch text {
// 		// 	case "exit":
// 		// 		io.WriteString(s, "GG\n")
// 		// 		return
// 		//
// 		// 	default:
// 		// 		io.WriteString(s, "You said:"+text)
// 		// 	}
// 		//
// 		// }
//
// 		// if err := scanner.Err(); err != nil {
// 		// 	log.Println("Error reading SSH input:", err.Error())
// 		// }
// 	})
//
// 	server := &ssh.Server{
// 		Addr: ":22",
// 		HostSigners: []ssh.Signer{
// 			privateKey,
// 		},
// 		PasswordHandler: func(ctx ssh.Context, password string) bool {
// 			// ctx.User()
// 			if password != "secret" {
// 				return false
// 			}
//
// 			return true
// 		},
// 	}
//
// 	// Start the server on port 2222
// 	log.Fatal(server.ListenAndServe())
// }
