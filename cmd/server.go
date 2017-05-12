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
	"fmt"

	"github.com/spf13/cobra"
	"net/http"
	"github.com/spf13/viper"
	"errors"
	"github.com/pupizoid/fatty/lib"
	"io/ioutil"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server for testing proxy",
	Long: ``,
	RunE: serve,
}


func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.Flags().String("ip", "127.0.0.1", "Listen ip address")
	serverCmd.Flags().Int("port", 3128, "Listen port")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func serve(c *cobra.Command, args []string) (err error) {

	var ip, addr string
	var port int

	if viper.ConfigFileUsed() != "" {
		port = viper.GetInt("server.port")
		ip = viper.GetString("server.ip")
	} else {
		if ip, err = c.Flags().GetString("ip"); err != nil {
			return
		}
		if port, err = c.Flags().GetInt("port"); err != nil {
			return
		}
	}

	switch {
	case ip == "*" && port != 0:
		addr = fmt.Sprintf(":%d", port)
	case ip != "" && port != 0:
		addr = fmt.Sprintf("%s:%d", ip, port)
	default:
		return errors.New("Unsupported ip & port combination")
	}

	http.HandleFunc("/", handler)
	fmt.Printf("Starting server on %s:%d\n", ip, port)
	if err = http.ListenAndServe(addr, nil); err != nil {
		fmt.Println(err)
		return
	}
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("header len: %d\n", len(r.Header.Get(lib.RequestHeaderName)))
	p, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	fmt.Printf("body len: %d\n", len(p))

	fmt.Printf("%#v\n", r)
	fmt.Printf("%#v\n", r.URL)

}