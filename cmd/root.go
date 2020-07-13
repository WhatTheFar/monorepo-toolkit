package cmd

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "monorepo-toolkit",
		Short: "A toolkit for monorepo",
		Long:  `monorepo-toolkit is a unified cli to manage your monorepo and CI/CD workflows`,
	}
	return rootCmd
}

// Execute executes the root command.
func Execute() error {
	return newMonorepoToolkit().
		Execute()
}

func newMonorepoToolkit() *cobra.Command {
	return newCommandsBuilder().
		addAll().
		build()
}

func newCommandsBuilder() *commandsBuilder {
	return &commandsBuilder{}
}

type commandsBuilder struct {
	commands []cmder
}

func (b *commandsBuilder) addAll() *commandsBuilder {
	b.addCommands(
		b.newListCmd(),
	)
	return b
}

func (b *commandsBuilder) addCommands(commands ...cmder) *commandsBuilder {
	b.commands = append(b.commands, commands...)
	return b
}

func (b *commandsBuilder) build() *cobra.Command {
	rootCmd := newRootCmd()
	addCommands(rootCmd, b.commands...)
	return rootCmd
}

func addCommands(root *cobra.Command, commands ...cmder) {
	for _, command := range commands {
		cmd := command.getCommand()
		if cmd == nil {
			continue
		}
		root.AddCommand(cmd)
	}
}

type cmder interface {
	getCommand() *cobra.Command
}

type baseCmd struct {
	cmd *cobra.Command
}

func (c *baseCmd) getCommand() *cobra.Command {
	return c.cmd
}
