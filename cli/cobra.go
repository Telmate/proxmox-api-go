package cli

import (
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

func Execute() {
    if err := RootCmd.Execute(); err != nil {
    	LogFatalError(err)
   	}
}