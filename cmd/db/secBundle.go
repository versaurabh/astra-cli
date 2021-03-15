//   Copyright 2021 Ryan Svihla
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

//Package db provides the sub-commands for the db command
package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rsds143/astra-cli/pkg"
	"github.com/rsds143/astra-cli/pkg/httputils"
	"github.com/rsds143/astra-devops-sdk-go/astraops"
	"github.com/spf13/cobra"
)

var secBundleFmt string
var secBundleLoc string

func init() {
	SecBundleCmd.Flags().StringVarP(&secBundleFmt, "output", "o", "zip", "Output format for report default is zip")
	SecBundleCmd.Flags().StringVarP(&secBundleLoc, "location", "l", "secureBundle.zip", "location of bundle to download to if using zip format. ignore if using json")
}

//SecBundleCmd  provides the secBundle database command
var SecBundleCmd = &cobra.Command{
	Use:   "secBundle <id>",
	Short: "get secure bundle by databaseID",
	Long:  `gets the secure connetion bundle for the database from your Astra account by ID`,
	Args:  cobra.ExactArgs(1),
	Run: func(cobraCmd *cobra.Command, args []string) {
		client, err := pkg.LoginClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to login with error %v\n", err)
			os.Exit(1)
		}
		id := args[0]
		var secBundle astraops.SecureBundle
		if secBundle, err = client.GetSecureBundle(id); err != nil {
			fmt.Fprintf(os.Stderr, "unable to get '%s' with error %v\n", id, err)
			os.Exit(1)
		}
		switch secBundleFmt {
		case "zip":
			httpClient := httputils.NewHTTPClient()
			res, err := httpClient.Get(secBundle.DownloadURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to download zip with error %v\n", err)
				os.Exit(1)
			}
			defer func() {
				err = res.Body.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warn: error closing http response body %v\n for request %v with status code %v", err, secBundle.DownloadURL, res.StatusCode)
				}
			}()
			f, err := os.Create(secBundleLoc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to create file to save too %v\n", err)
				os.Exit(1)
			}
			defer func() {
				err = f.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warn: error closing file %v for file %v\n", err, secBundleLoc)
				}
			}()
			i, err := io.Copy(f, res.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to copy downloaded file to %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("file %v saved %v bytes written\n", secBundleLoc, i)
		case "json":
			b, err := json.MarshalIndent(secBundle, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "unexpected error marshaling to json: '%v', Try -output text instead\n", err)
				os.Exit(1)
			}
			fmt.Println(string(b))
		default:
			fmt.Fprintf(os.Stderr, "-output %q is not valid option.", secBundleFmt)
			os.Exit(1)
		}
	},
}