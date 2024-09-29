package cli

import (
	"fmt"
	"os"

	"github.com/bedirhangull/kubectl-crashwatch/pkg/plugin"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
	dashboardFlag         bool
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl-crashwatch",
		Short: "Display logs of CrashLoopBackOff pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dashboardFlag {
				plugin.RunDashboard(KubernetesConfigFlags)
				return nil
			}

			if err := plugin.RunPlugin(KubernetesConfigFlags); err != nil {
				fmt.Println(err)
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&dashboardFlag, "dashboard", "d", false, "Display the dashboard")

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
