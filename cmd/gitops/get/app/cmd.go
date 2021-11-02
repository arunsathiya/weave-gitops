package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

var Cmd = &cobra.Command{
	Use:     "app",
	Aliases: []string{"apps"},
	Short:   "List applications under gitops control",
	Example: "gitops get apps",
	RunE:    runCmd,
}

func runCmd(cmd *cobra.Command, args []string) error {
	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error initializing kubernetes client: %w", err)
	}

	ns, err := cmd.Parent().Parent().Flags().GetString("namespace")
	if err != nil {
		return err
	}

	apps, err := kubeClient.GetApplications(context.Background(), ns)
	if err != nil {
		return err
	}

	fmt.Println("NAME")

	for _, app := range apps {
		fmt.Println(app.Name)
	}

	return nil
}