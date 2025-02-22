package cluster

import (
	"context"
	"fmt"
	"log"
	"os"

	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift/osdctl/internal/utils/globalflags"
	"github.com/openshift/osdctl/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const loggingLabel string = "ext-managed.openshift.io/extended-logging-support"

// loggingCheckOptions defines the struct for running loggingCheck command
// This command requires the ocm API Token https://cloud.redhat.com/openshift/token be available in the OCM_TOKEN env variable.

type loggingCheckOptions struct {
	output    string
	verbose   bool
	clusterID string

	genericclioptions.IOStreams
	GlobalOptions *globalflags.GlobalOptions
}

// newCmdloggingCheck implements the loggingCheck command to show the logging support status of a cluster
func newCmdloggingCheck(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags, globalOpts *globalflags.GlobalOptions) *cobra.Command {
	ops := newloggingCheckOptions(streams, flags, globalOpts)
	loggingCheckCmd := &cobra.Command{
		Use:               "loggingCheck",
		Short:             "Shows the logging support status of a specified cluster",
		Args:              cobra.ExactArgs(1),
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(ops.complete(cmd, args))
			cmdutil.CheckErr(ops.run())
		},
	}
	loggingCheckCmd.Flags().BoolVarP(&ops.verbose, "verbose", "", false, "Verbose output")

	return loggingCheckCmd
}

func newloggingCheckOptions(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags, globalOpts *globalflags.GlobalOptions) *loggingCheckOptions {
	return &loggingCheckOptions{
		IOStreams:     streams,
		GlobalOptions: globalOpts,
	}
}

func (o *loggingCheckOptions) complete(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return cmdutil.UsageErrorf(cmd, "Provide exactly one cluster ID")
	}

	o.clusterID = args[0]
	o.output = o.GlobalOptions.Output

	return nil
}

func (o *loggingCheckOptions) run() error {

	// Create a context:
	ctx := context.Background()
	//The ocm
	token := os.Getenv("OCM_TOKEN")
	if token == "" {
		ocmToken, err := utils.GetOCMAccessToken()
		if err != nil {
			log.Fatalf("OCM token not set. Please configure it using the OCM_TOKEN evnironment variable or the ocm cli")
			os.Exit(1)
		}
		token = *ocmToken
	}
	connection, err := sdk.NewConnectionBuilder().
		Tokens(token).
		Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't build connection: %v\n", err)
		os.Exit(1)
	}
	defer connection.Close()

	// Get the client for the resource that manages the collection of clusters:
	collection := connection.ClustersMgmt().V1().Clusters()
	// Get the labels externally available for the cluster
	resource := collection.Cluster(o.clusterID).ExternalConfiguration().Labels()
	// Send the request to retrieve the list of external cluster labels:
	response, err := resource.List().SendContext(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't retrieve cluster labels: %v\n", err)
		os.Exit(1)
	}

	labels, ok := response.GetItems()
	// If there are no labels, then logging is not SREP supported
	if !ok {
		fmt.Printf("Cluster logging not SREP supported\n")
		return nil
	}

	for _, label := range labels.Slice() {
		if l, ok := label.GetKey(); ok {
			// If the label is found as the key, we know its an SREP supported logging stack
			if l == loggingLabel {
				fmt.Printf("Cluster logging SREP supported for the target cluster\n")
				return nil
			}
		}
	}

	fmt.Printf("Cluster logging not SREP supported\n")

	return nil
}
