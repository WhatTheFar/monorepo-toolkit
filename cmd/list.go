package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/whatthefar/monorepo-toolkit/pkg/factory"
)

func newListCmdFlag() *listCmdFlag {
	f := &listCmdFlag{}
	viper.Unmarshal(f)
	return f
}

type listCmdFlag struct {
	CITool     string `mapstructure:"ciTool"`
	WorkflowID string `mapstructure:"workflowID"`
}

func newListProjectsCmdFlag() *listProjectsCmdFlag {
	f := &listProjectsCmdFlag{}
	viper.Unmarshal(f)
	return f
}

type listProjectsCmdFlag struct {
	listCmdFlag `mapstructure:",squash"`
	Join        bool `mapstructure:"join"`
}

func (f *listCmdFlag) validate() error {
	missing := make([]string, 0)
	if f.CITool == "" {
		missing = append(missing, "CI_TOOL")
	}
	if f.WorkflowID == "" {
		missing = append(missing, "WORKFLOW_ID")
	}
	if len(missing) > 0 {
		for i, v := range missing {
			missing[i] = fmt.Sprintf(`"%s"`, v)
		}
		joined := strings.Join(missing, ", ")
		return errors.Errorf("required flags(s) %s not set", joined)
	}
	return nil
}

var (
	listCmd         *cobra.Command
	listProjectsCmd *cobra.Command

	ciControllerFactory = factory.CIController
)

func (b *commandsBuilder) newListCmd() *baseCmd {
	listCmd = &cobra.Command{
		Use:   "list [command]",
		Short: "Listing out various types of information",
		Long: `Listing out varios types of information.

List requires a subcommand, e.g., ` + "`monorepo-toolkit list projects`",
	}

	listProjectsCmd = &cobra.Command{
		Use:   "projects [paths]",
		Short: "Listing projects that have changes",
		Long:  `Listing projects that have changes since the last successful build`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			f := newListProjectsCmdFlag()
			err := f.validate()
			if err != nil {
				er(errors.Wrap(err, "fail to validate flags for list command"))
			}

			workDir, err := os.Getwd()
			if err != nil {
				er(errors.Wrap(err, "can't get current directory"))
			}
			ctrl, err := ciControllerFactory.New(workDir, f.CITool)
			if err != nil {
				er(errors.Wrap(err, "can't create CI controller"))
			}

			ctx := context.Background()
			if f.Join == true {
				ctrl.ListProjectsJoined(ctx, args, f.WorkflowID)
			} else {
				ctrl.ListProjects(ctx, args, f.WorkflowID)
			}
		},
	}

	listCmd.PersistentFlags().StringP("ci-tool", "C", "", `CI provider, e.g., "github"`)
	listCmd.PersistentFlags().StringP("workflow", "W", "", "Workflow ID, e.g., a file name for github action")
	viper.BindPFlag("ciTool", listCmd.PersistentFlags().Lookup("ci-tool"))
	viper.BindPFlag("workflowID", listCmd.PersistentFlags().Lookup("workflow"))
	viper.BindEnv("ciTool", "CI_TOOL")
	viper.BindEnv("workflowID", "WORKFLOW_ID")

	listProjectsCmd.Flags().Bool("join", false, `join projects into single project (default false)`)
	viper.BindPFlag("join", listProjectsCmd.Flags().Lookup("join"))

	listCmd.AddCommand(listProjectsCmd)

	return &baseCmd{cmd: listCmd}
}
