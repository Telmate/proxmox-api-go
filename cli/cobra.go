package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

// Global else the nested folders don't work
var RootCmd = &cobra.Command{
	Use:   "proxmox-api-go",
	Short: "Application to configure Proxmox from the Api",
}

func init() {
	RootCmd.PersistentFlags().BoolP("insecure", "i", false, "TLS insecure mode")
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")
	RootCmd.PersistentFlags().IntP("timeout", "t", 300, "api task timeout in seconds")
	RootCmd.PersistentFlags().StringP("file", "f", "", "file to get the config from")
	RootCmd.PersistentFlags().StringP("proxyurl", "p", "", "proxy url to connect to")
}

func Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	// Channel to catch OS signals
	signalChan := make(chan os.Signal, 1)

	// Notify signalChan when SIGINT or SIGTERM is received
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Goroutine to handle signal
	go func() {
		defer signal.Stop(signalChan) // Cleanup when done
		<-signalChan                  // Wait for a signal
		cancel()                      // Cancel the context
	}()
	return ctx
}

func Execute() (err error) {
	if err = RootCmd.Execute(); err != nil {
		return
	}
	return
}

func NewClient() (c *proxmox.Client) {
	c, err := Client(Context(), "", "", "", "", "")
	LogFatalError(err)
	return
}

func Client(ctx context.Context, apiUrl, userID, password, otp string, http_headers string) (c *proxmox.Client, err error) {
	insecure, _ := RootCmd.Flags().GetBool("insecure")
	timeout, _ := RootCmd.Flags().GetInt("timeout")
	proxyUrl, _ := RootCmd.Flags().GetString("proxyurl")

	tlsConf := &tls.Config{InsecureSkipVerify: true}
	if !insecure {
		tlsConf = nil
	}
	if apiUrl == "" {
		apiUrl = os.Getenv("PM_API_URL")
	}
	if userID == "" {
		userID = os.Getenv("PM_USER")
	}
	if password == "" {
		password = os.Getenv("PM_PASS")
	}
	if otp == "" {
		otp = os.Getenv("PM_OTP")
	}
	if http_headers == "" {
		http_headers = os.Getenv("PM_HTTP_HEADERS")
	}
	c, err = proxmox.NewClient(apiUrl, nil, http_headers, tlsConf, proxyUrl, timeout, false)
	LogFatalError(err)
	if userRequiresAPIToken(userID) {
		var token proxmox.ApiTokenID
		LogFatalError(token.Parse(userID))
		c.SetAPIToken(token, proxmox.ApiTokenSecret(password))
		// As test, get the version of the server
		_, err = c.GetVersion(Context())
		if err != nil {
			err = fmt.Errorf("login error: %s", err)
		}
	} else {
		err = c.Login(ctx, userID, password, otp)
	}
	return
}

var rxUserRequiresToken = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+![a-z0-9]+")

func userRequiresAPIToken(userID string) bool {
	return rxUserRequiresToken.MatchString(userID)
}

func NewConfig() (configSource []byte) {
	var err error
	file, _ := RootCmd.Flags().GetString("file")
	if file != "" {
		configSource, err = os.ReadFile(file)
		LogFatalError(err)
	} else {
		configSource, err = io.ReadAll(RootCmd.InOrStdin())
		LogFatalError(err)
	}
	return
}
