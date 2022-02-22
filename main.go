package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/mywordoftheday/backend/internal/mail"
	"github.com/mywordoftheday/backend/internal/server"
	v1alpha1 "github.com/mywordoftheday/proto/mywordoftheday/v1alpha1"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	logrus.SetOutput(os.Stdout)

	// Only log the info severity or above.
	logrus.SetLevel(logrus.InfoLevel)
}

func handleBindEnvErr(err error) {
	if err != nil {
		logrus.Fatalf("unable to bind viper key to environment variable: '%+v'", err)
	}
}

func main() {
	// Config files
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Custom config file mapped as a volume when using Docker
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/config")

	// Environment Variables
	handleBindEnvErr(viper.BindEnv("server.port", "SERVER_PORT"))
	handleBindEnvErr(viper.BindEnv("server.httpProxy.enabled", "HTTP_PROXY_ENABLED"))
	handleBindEnvErr(viper.BindEnv("server.httpProxy.port", "HTTP_PROXY_PORT"))

	handleBindEnvErr(viper.BindEnv("db.host", "DB_HOST"))
	handleBindEnvErr(viper.BindEnv("db.port", "DB_PORT"))
	handleBindEnvErr(viper.BindEnv("db.username", "DB_USERNAME"))
	handleBindEnvErr(viper.BindEnv("db.password", "DB_PASSWORD"))
	handleBindEnvErr(viper.BindEnv("db.name", "DB_NAME"))

	handleBindEnvErr(viper.BindEnv("smtp.enabled", "SMTP_ENABLED"))
	handleBindEnvErr(viper.BindEnv("smtp.schedule", "SMTP_SCHEDULE"))
	handleBindEnvErr(viper.BindEnv("smtp.host", "SMTP_HOST"))
	handleBindEnvErr(viper.BindEnv("smtp.port", "SMTP_PORT"))
	handleBindEnvErr(viper.BindEnv("smtp.username", "SMTP_USERNAME"))
	handleBindEnvErr(viper.BindEnv("smtp.password", "SMTP_PASSWORD"))
	handleBindEnvErr(viper.BindEnv("smtp.fromAddress", "SMTP_FROM_ADDRESS"))
	handleBindEnvErr(viper.BindEnv("smtp.toAddresses", "SMTP_TO_ADDRESSES"))

	// Merge config
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore as we use defaults/environment variables
			// and if anything required isn't set (e.g. db password) we'll error later on
		} else {
			// Config file was found but another error was produced
			logrus.Fatalf("unable to merge in config: '%+v'", err)
		}
	}

	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.httpProxy.enabled", false)
	viper.SetDefault("server.httpProxy.port", 8443)

	// DB defaults
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.username", "mywordoftheday")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.name", "mywordoftheday")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore as we use defaults/environment variables
			// and if anything required isn't set (e.g. db password) we'll error later on
		} else {
			// Config file was found but another error was produced
			logrus.Fatal("unable to read config: ", err)
		}
	}

	var (
		port             = viper.GetInt("server.port")
		httpProxyEnabled = viper.GetBool("server.httpProxy.enabled")
		httpProxyPort    = viper.GetInt("server.httpProxy.port")

		dbHost     = viper.GetString("db.host")
		dbPort     = viper.GetString("db.port")
		dbUsername = viper.GetString("db.username")
		dbPassword = viper.GetString("db.password")
		dbName     = viper.GetString("db.name")

		smtpEnabled     = viper.GetBool("smtp.enabled")
		smtpSchedule    = viper.GetString("smtp.schedule")
		smtpHost        = viper.GetString("smtp.host")
		smtpPort        = viper.GetString("smtp.port")
		smtpUsername    = viper.GetString("smtp.username")
		smtpPassword    = viper.GetString("smtp.password")
		smtpFromAddress = viper.GetString("smtp.fromAddress")
		smtpToAddresses = viper.GetStringSlice("smtp.toAddresses")
	)

	logrus.WithFields(logrus.Fields{
		"Server Port":        port,
		"HTTP Proxy Enabled": httpProxyEnabled,
		"HTTP Proxy Port":    httpProxyPort,
		"Database Name":      dbName,
		"Database Host":      dbHost,
		"Database Port":      dbPort,
		"Database Username":  dbUsername,
		"SMTP Enabled":       smtpEnabled,
		"SMTP Schedule":      smtpSchedule,
	}).Info("Config Initialised")

	svr, err := server.New(
		server.Config{DBHost: dbHost, DBPort: dbPort, DBUsername: dbUsername, DBPassword: dbPassword, DBName: dbName},
	)
	if err != nil {
		logrus.Fatalf("Unable to initialise new Server: %+v", err)
	}

	gServer := grpc.NewServer()

	v1alpha1.RegisterMyWordOfTheDayServiceServer(gServer, svr)

	reflection.Register(gServer)

	addr := fmt.Sprintf(":%d", port)

	if httpProxyEnabled {
		go httpProxyServer(httpProxyPort, addr)
	}

	if smtpEnabled {
		mailClient, err := mail.New(mail.Config{
			SMTPHost:        smtpHost,
			SMTPPort:        smtpPort,
			SMTPUsername:    smtpUsername,
			SMTPPassword:    smtpPassword,
			SMTPFromAddress: smtpFromAddress,
			SMTPToAddresses: smtpToAddresses,
		}, "template.html")
		if err != nil {
			log.Fatalf("Error creating new mail client: %+v", err)
		}

		// Parse the SMTP Schedule and make sure it's valid
		if _, err := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(smtpSchedule); err != nil {
			log.Fatalf("Error parsing smtp schedule: %+v", err)
		}

		c := cron.New()
		c.AddFunc(smtpSchedule, func() {
			rw, err := svr.RandomWord(context.Background(), &v1alpha1.RandomWordRequest{})
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("Error getting random word")
			}

			if rw.GetWord().GetId() == 0 && rw.GetWord().GetWord() == "" && rw.GetWord().GetCustomDefinition() == "" {
				logrus.Info("No words have been added - skipping")
				return
			}

			if err := mailClient.SendMailFromTemplate("My Word Of The Day", struct {
				Word       string
				Definition string
			}{
				Word:       rw.GetWord().GetWord(),
				Definition: rw.GetWord().GetCustomDefinition(),
			}); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("Error sending mail")
			}
		})

		c.Start()
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Fatal(err, "Failed to create listener")
	}

	logrus.WithFields(logrus.Fields{
		"port": port,
	}).Info("Starting grpc server")

	if err := gServer.Serve(listener); err != nil {
		logrus.Fatal(err, "Failed to start server")
	}
}

// httpProxyServer starts a new http server listening on the specified port, proxying
// requests to the provided grpc service
func httpProxyServer(port int, grpcAddr string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	grpcMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if err := v1alpha1.RegisterMyWordOfTheDayServiceHandlerFromEndpoint(ctx, grpcMux, grpcAddr, opts); err != nil {
		logrus.Fatal(err, "Failed to register http handler")
	}

	r := http.NewServeMux()

	r.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// gateway is generated to match for /v1alpha1/ and not /api/v1alpha1
		// we could update the gateway proto to match for /api/v1alpha1 but
		// it shouldn't care where it's mounted to, hence we just rewrite the path here
		r.URL.Path = strings.Replace(r.URL.Path, "/api", "", -1)
		grpcMux.ServeHTTP(w, r)
	})

	logrus.WithFields(logrus.Fields{
		"port": port,
	}).Info("Starting http proxy server")

	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r), "Failed to start http proxy server")
}
