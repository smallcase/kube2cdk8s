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
			multiple := viper.GetBool("multiple")

			var result string

			if filePath == "" {
				log.Fatal("-f, --file is required")
			}

			if multiple {
				result, err := kube2cdk8s.Kube2CDK8SMultiple(filePath)
				if err != nil {
					return err
				}

				fmt.Print(result)
				return nil
			}

			result, err := kube2cdk8s.Kube2CDK8S(filePath)
			if err != nil {
				return err
			}

			fmt.Print(result)
			return nil
		}}

	return command
}
