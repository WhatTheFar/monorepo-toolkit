package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newBuildCmdFlag() *buildCmdFlag {
	f := &buildCmdFlag{}
	buildCmdViper.Unmarshal(f)
	return f
}

type buildCmdFlag struct {
	CITool     string `mapstructure:"ciTool"`
	WorkflowID string `mapstructure:"workflowID"`
	Once       bool   `mapstructure:"once"`
}

func (f *buildCmdFlag) validate() error {
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
	buildCmd *cobra.Command

	buildCmdViper = viper.New()
)

func (b *commandsBuilder) newBuildCmd() *baseCmd {
	buildCmd = &cobra.Command{
		Use:   "build [paths]",
		Short: "Trigger build workflow for projects that have changes",
		Long:  "Trigger build workflow for projects that have changes",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			f := newBuildCmdFlag()
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
			if f.Once == true {
				ctrl.BuildOnce(ctx, args, f.WorkflowID)
			} else {
				ctrl.Build(ctx, args, f.WorkflowID)
			}
		},
	}

	buildCmd.PersistentFlags().StringP("ci-tool", "C", "", `CI provider, e.g., "github"`)
	buildCmd.PersistentFlags().StringP("workflow", "W", "", "Workflow ID, e.g., a file name for github action")
	buildCmdViper.BindPFlag("ciTool", buildCmd.PersistentFlags().Lookup("ci-tool"))
	buildCmdViper.BindPFlag("workflowID", buildCmd.PersistentFlags().Lookup("workflow"))
	buildCmdViper.BindEnv("ciTool", "CI_TOOL")
	buildCmdViper.BindEnv("workflowID", "WORKFLOW_ID")

	buildCmd.Flags().Bool("once", false, `join projects into single project and trigger only one workflow (default false)`)
	buildCmdViper.BindPFlag("once", buildCmd.Flags().Lookup("once"))

	return &baseCmd{cmd: buildCmd}
}
