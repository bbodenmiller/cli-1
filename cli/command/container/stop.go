package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/completion"
	"github.com/docker/docker/api/types/container"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type stopOptions struct {
	time        int
	timeChanged bool

	containers []string
}

// NewStopCommand creates a new cobra.Command for `docker stop`
func NewStopCommand(dockerCli command.Cli) *cobra.Command {
	var opts stopOptions

	cmd := &cobra.Command{
		Use:   "stop [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "Stop one or more running containers",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.containers = args
			opts.timeChanged = cmd.Flags().Changed("time")
			return runStop(dockerCli, &opts)
		},
		ValidArgsFunction: completion.ContainerNames(dockerCli, false),
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.time, "time", "t", 10, "Seconds to wait for stop before killing it")
	return cmd
}

func runStop(dockerCli command.Cli, opts *stopOptions) error {
	var timeout *int
	if opts.timeChanged {
		timeout = &opts.time
	}

	errChan := parallelOperation(context.Background(), opts.containers, func(ctx context.Context, id string) error {
		return dockerCli.Client().ContainerStop(ctx, id, container.StopOptions{
			Timeout: timeout,
		})
	})
	var errs []string
	for _, ctr := range opts.containers {
		if err := <-errChan; err != nil {
			errs = append(errs, err.Error())
			continue
		}
		_, _ = fmt.Fprintln(dockerCli.Out(), ctr)
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
