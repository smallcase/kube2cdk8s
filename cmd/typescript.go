package cmd

import (
	"fmt"
	"log"

	"github.com/smallcase/kube2cdk8s/pkg/kube2cdk8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TSCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "typescript",
		Long: "convert k8s yaml to typescript",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := viper.GetString("file")

			if filePath == "" {
				log.Fatal("-f, --file is required")
			}

			data, err := kube2cdk8s.Kube2CDK8S(filePath)
			if err != nil {
				return err
			}

			fmt.Print(data)
			return nil
		}}

	return command
}
