package cli

import (
	"crypto/tls"
	"log"
	"os"
	"regexp"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

// Global else the nested folders dont work
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

func Execute() (err error) {
    if err = RootCmd.Execute(); err != nil {
    	return
   	}
	return
}

func NewClient()(c *proxmox.Client) {
	insecure, _ := RootCmd.Flags().GetBool("insecure")
	timeout, _ := RootCmd.Flags().GetInt("timeout")
	proxyUrl, _ := RootCmd.Flags().GetString("proxyurl")

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !insecure {
		tlsconf = nil
	}

	c, err := proxmox.NewClient(os.Getenv("PM_API_URL"), nil, tlsconf, proxyUrl, timeout)
	LogFatalError(err)
	if userRequiresAPIToken(os.Getenv("PM_USER")) {
		c.SetAPIToken(os.Getenv("PM_USER"), os.Getenv("PM_PASS"))
		// As test, get the version of the server
		_, err := c.GetVersion()
		if err != nil {
			log.Fatalf("login error: %s", err)
		}
	} else {
		err := c.Login(os.Getenv("PM_USER"), os.Getenv("PM_PASS"), os.Getenv("PM_OTP"))
		LogFatalError(err)
	}
	return
}

var rxUserRequiresToken = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+![a-z0-9]+")

func userRequiresAPIToken(userID string) bool {
	return rxUserRequiresToken.MatchString(userID)
}
