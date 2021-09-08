/*
Copyright © 2021 smallcase infra <infra@smallcase.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/smallcase/kube2cdk8s/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	manifestFile string
	multiple     bool
)

func configureCLI() *cobra.Command {
	rootCmd := &cobra.Command{Use: "kube2cdk8s", Long: "converts k8s yaml to cdk8s"}

	rootCmd.AddCommand(cmd.TSCommand())

	rootCmd.PersistentFlags().StringVarP(&manifestFile, "file", "f", "", "YAML file to convert")
	err := viper.BindPFlag("file", rootCmd.PersistentFlags().Lookup("file"))
	if err != nil {
		log.Println(err)
	}

	rootCmd.PersistentFlags().BoolVarP(&multiple, "multiple", "m", false, "convert multiple yamls seperated by ---")
	err = viper.BindPFlag("multiple", rootCmd.PersistentFlags().Lookup("multiple"))
	if err != nil {
		log.Println(err)
	}

	return rootCmd
}

func main() {
	rootCmd := configureCLI()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("unable to run program: %v\n", err)
		os.Exit(1)
	}
}
