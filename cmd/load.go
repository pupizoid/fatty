package cmd

import (
	"github.com/spf13/cobra"
	"github.com/pupizoid/fatty/lib"
	"os"
	"net/url"
	"strings"
	"encoding/xml"
	"io/ioutil"
	"fmt"
)

type Result struct {
	For For `xml:"for"`
}

type For struct {
	Request []Request `xml:"request"`
}

type Request struct {
	Http Http `xml:"http"`
}

type Http struct {
	Url     string `xml:"url,attr"`
	Version string `xml:"version,attr"`
	Method  string `xml:"method,attr"`
}

// loadDmd represents the test command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Runs a test of desired web server or proxy",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		list, err := cmd.Flags().GetString("list")
		workers, err := cmd.Flags().GetUint("workers")
		timeout, err := cmd.Flags().GetInt("timeout")
		ip, err := cmd.Flags().GetString("ip")
		port, err := cmd.Flags().GetString("port")

		//proxy, err := cmd.Flags().GetString("proxy")
		proxyUser, err := cmd.Flags().GetString("proxy-user")
		proxyPass, err := cmd.Flags().GetString("proxy-pass")
		if err != nil {
			return err
		}

		disp := lib.NewDispatcher(timeout)

		options := lib.LoadEmitterOptions{
			Ip:   ip,
			Port: port,
		}

		var ps *url.URL
		if !strings.Contains(ip, "http://") && !strings.Contains(ip, "https://") {
			ip = "http://" + ip // support only http proxy for now
		}

		ps, err = lib.ParseProxy(fmt.Sprintf("%s:%s", ip, port))
		if err != nil {
			return err
		}

		if proxyUser != "" {
			if proxyPass != "" {
				ps.User = url.UserPassword(proxyUser, proxyPass)
			} else {
				ps.User = url.User(proxyUser)
			}
		}

		if f, err := os.Open(list); err == nil {
			var s Result
			b, _ := ioutil.ReadAll(f)
			xml.Unmarshal(b, &s)
			options.Urls = make(chan string, len(s.For.Request))
			for _, req := range s.For.Request {
				options.Urls <- req.Http.Url
			}
			f.Close()
		} else {
			return err
		}

		for i := uint(0); i < workers; i++ {
			emitter := lib.NewLoadEmitter(&options, ps)
			disp.Emitters = append(disp.Emitters, emitter)
		}

		disp.Run()
		return
	},
}

func init() {
	RootCmd.AddCommand(loadCmd)

	loadCmd.Flags().StringP("list", "l", "", "Path to file with url list")
	loadCmd.Flags().UintP("workers", "w", 1, "Number of concurrent requests")
	loadCmd.Flags().IntP("timeout", "t", 0, "Maximum test duration(0=endless)")
	loadCmd.Flags().StringP("ip", "i", "", "Destination IP address")
	loadCmd.Flags().StringP("port", "p", "8080", "Destination port")

	loadCmd.Flags().String("proxy", "", "Proxy server url. Can contain basic proxy authentication.")
	loadCmd.Flags().String("proxy-user", "", "Proxy user login")
	loadCmd.Flags().String("proxy-pass", "", "Proxy user password")

}
