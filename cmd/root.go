package cmd

import (
	routev1 "github.com/openshift/api/route/v1"
	awsv1alpha1 "github.com/openshift/aws-account-operator/pkg/apis/aws/v1alpha1"
	gcpv1alpha1 "github.com/openshift/gcp-project-operator/pkg/apis"
	hivev1 "github.com/openshift/hive/pkg/apis/hive/v1"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/openshift/osdctl/cmd/aao"
	"github.com/openshift/osdctl/cmd/account"
	"github.com/openshift/osdctl/cmd/cluster"
	"github.com/openshift/osdctl/cmd/clusterdeployment"
	"github.com/openshift/osdctl/cmd/cost"
	"github.com/openshift/osdctl/cmd/federatedrole"
	"github.com/openshift/osdctl/cmd/network"
	"github.com/openshift/osdctl/cmd/servicelog"
	"github.com/openshift/osdctl/cmd/sts"
	"github.com/openshift/osdctl/internal/utils/globalflags"
	"github.com/openshift/osdctl/pkg/k8s"
)

func init() {
	_ = awsv1alpha1.AddToScheme(scheme.Scheme)
	_ = routev1.AddToScheme(scheme.Scheme)
	_ = hivev1.AddToScheme(scheme.Scheme)
	_ = gcpv1alpha1.AddToScheme(scheme.Scheme)
}

// NewCmdRoot represents the base command when called without any subcommands
func NewCmdRoot(streams genericclioptions.IOStreams) *cobra.Command {
	globalOpts := &globalflags.GlobalOptions{}
	rootCmd := &cobra.Command{
		Use:               "osdctl",
		Short:             "OSD CLI",
		Long:              `CLI tool to provide OSD related utilities`,
		DisableAutoGenTag: true,
		Run:               help,
	}

	globalflags.AddGlobalFlags(rootCmd, globalOpts)
	kubeFlags := globalflags.GetFlags(rootCmd)

	kubeClient := k8s.NewClient(kubeFlags)

	// add sub commands
	rootCmd.AddCommand(aao.NewCmdAao(streams, kubeFlags))
	rootCmd.AddCommand(account.NewCmdAccount(streams, kubeFlags, kubeClient, globalOpts))
	rootCmd.AddCommand(cluster.NewCmdCluster(streams, kubeFlags, globalOpts))
	rootCmd.AddCommand(clusterdeployment.NewCmdClusterDeployment(streams, kubeFlags, kubeClient))
	rootCmd.AddCommand(federatedrole.NewCmdFederatedRole(streams, kubeFlags, kubeClient))
	rootCmd.AddCommand(network.NewCmdNetwork(streams, kubeFlags, kubeClient))
	rootCmd.AddCommand(newCmdMetrics(streams, kubeFlags, kubeClient))
	rootCmd.AddCommand(servicelog.NewCmdServiceLog())
	rootCmd.AddCommand(sts.NewCmdSts(streams, kubeFlags, kubeClient))

	// add docs command
	rootCmd.AddCommand(newCmdDocs(streams))

	// add completion command
	rootCmd.AddCommand(newCmdCompletion(streams))

	// add options command to list global flags
	templates.ActsAsRootCommand(rootCmd, []string{"options"})
	rootCmd.AddCommand(newCmdOptions(streams))

	// Add cost command to use AWS Cost Manager
	rootCmd.AddCommand(cost.NewCmdCost(streams, globalOpts))

	// Add version subcommand. Using the in-build --version flag does not work with cobra
	// because there is no way to hook a function to the --version flag in cobra.
	rootCmd.AddCommand(versionCmd)

	// Add upgradeCmd for upgrading the currently running executable in-place.
	rootCmd.AddCommand(upgradeCmd)

	return rootCmd
}

func help(cmd *cobra.Command, _ []string) {
	cmd.Help()
}
