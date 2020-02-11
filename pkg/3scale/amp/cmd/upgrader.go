package cmd

import (
	"fmt"
	"os"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/cmd/upgrader"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade your 3scale installation",
	Long:  "Upgrade your 3scale installation",
	Run:   upgradeCommandEntrypoint,
}

func upgradeCommandEntrypoint(cmd *cobra.Command, args []string) {
	configuration := config.GetConfigOrDie()
	cl, err := client.New(configuration, client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	err = upgrader.CheckCurrentInstallation(cl, upgradeNamespace)
	if err != nil {
		fmt.Printf("failed to check current 3scale installation: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.MigrateSystemSMTPData(cl, upgradeNamespace)
	if err != nil {
		fmt.Printf("failed to migrate System SMTP data: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.UpgradeSystemPreHook(cl, upgradeNamespace)
	if err != nil {
		fmt.Printf("failed to migrate System SMTP data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("3scale successfully upgraded to release 2.8")
}

var upgradeNamespace string

func init() {
	upgradeCmd.Flags().StringVarP(&upgradeNamespace, "namespace", "n", "", "Cluster namespace (required)")
	upgradeCmd.MarkFlagRequired("namespace")
	rootCmd.AddCommand(upgradeCmd)

}
