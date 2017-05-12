// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"net/http"
	"github.com/pupizoid/fatty/lib"
	"net/url"
	"strings"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs a test of desired web server or proxy",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		dest, err := cmd.Flags().GetString("dest")
		workers, err := cmd.Flags().GetUint("workers")
		limit, err := cmd.Flags().GetUint32("limit")
		method, err := cmd.Flags().GetString("method")

		headerSize, err := cmd.Flags().GetUint("header-size")
		headerInc, err := cmd.Flags().GetUint("header-inc-rate")
		headerMulti, err := cmd.Flags().GetUint("header-multi-rate")

		bodySize, err := cmd.Flags().GetUint("body-size")
		bodyInc, err := cmd.Flags().GetUint("body-inc-rate")
		bodyMulti, err := cmd.Flags().GetUint("body-multi-rate")
		bodyFile, err := cmd.Flags().GetString("body-from-file")

		proxy, err := cmd.Flags().GetString("proxy")
		proxyUser, err := cmd.Flags().GetString("proxy-user")
		proxyPass, err := cmd.Flags().GetString("proxy-pass")

		if err != nil {return}

		disp := lib.NewDispatcher()

		ds, err := url.Parse(dest)
		if err != nil {
			return
		}


		var header lib.GrowableContent
		if headerSize > 0 {
			header = lib.NewContent(headerSize, headerInc, headerMulti)
		}

		var body lib.GrowableContent
		if bodyFile == "" {
			if bodySize > 0 {
				body = lib.NewContent(bodySize, bodyInc, bodyMulti)
			}
			// body = nil
		} else {
			body, err = lib.NewBodyFromFile(bodyFile)
			if err != nil {
				return
			}
		}

		var ps *url.URL
		if proxy != "" {
			if !strings.Contains(proxy, "http://") && !strings.Contains(proxy, "https://") {
				proxy = "http://" + proxy // support only http proxy for now
			}

			ps, err = lib.ParseProxy(proxy)
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
		}

		for i := 0; uint(i) < workers; i++ {
			disp.Emitters = append(disp.Emitters, lib.NewEmitter(
				method, limit, header, body, ds, ps,
			))
		}

		disp.Run()

		return
	},
}

func init() {
	RootCmd.AddCommand(testCmd)

	testCmd.Flags().StringP("dest", "d", "", "Requests destination")
	testCmd.Flags().UintP("workers", "w", 1, "Workers count (async testing)")
	testCmd.Flags().Uint32P("limit", "l", 0, "Count of requests to be sent, 0 = unlimited")
	testCmd.Flags().StringP("method", "m", http.MethodGet, "Request method")

	testCmd.Flags().Uint("header-size", 0, "Request header size (bytes)")
	testCmd.Flags().Uint("header-inc-rate", 0, "Request header amplification rate (bytes)")
	testCmd.Flags().Uint("header-multi-rate", 1, "Request header multiplication rate")

	testCmd.Flags().UintP("body-size", "b", 0, "Request body size (bytes)")
	testCmd.Flags().Uint("body-inc-rate", 0, "Request body amplification rate (bytes)")
	testCmd.Flags().Uint("body-multi-rate", 1, "Request body multiplication rate")
	testCmd.Flags().String("body-from-file", "", "Read request body content from file")

	testCmd.Flags().StringP("proxy", "p", "", "Proxy server url. Can contain basic proxy authentication.")
	testCmd.Flags().String("proxy-user", "", "Proxy user login")
	testCmd.Flags().String("proxy-pass", "", "Proxy user password")
}
