package repos

import (
	"context"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/zkhvan/gh-clone/internal/cmdutil"
)

type Options struct {
	Directory    string
	GitArgs      []string
	UpstreamName string

	Owner   string
	Workers int
}

func NewCmdRepos(runF func(*Options) error) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,

		Use:   "repos <directory> [-- <gitflags>...]",
		Args:  cmdutil.MinimumArgs(1, "cannot clone repos: directory argument required"),
		Short: "Clones multiple repositories locally.",
		Long: heredoc.Docf(`
			Clones multiple GitHub repositories locally. Pass additional %[1]sgit
			clone%[1]s flags by listing them after "--".
		`, "`"),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Directory = args[0]
			opts.GitArgs = args[1:]

			if runF != nil {
				return runF(opts)
			}

			return reposRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.UpstreamName, "upstream-remote-name", "u", "upstream", "Upstream remote name when cloning a fork")
	cmd.Flags().StringVarP(&opts.Owner, "owner", "o", "", "Repository owner")
	cmd.Flags().IntVar(&opts.Workers, "workers", 10, "Number of workers to spawn for cloning")
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err == pflag.ErrHelp {
			return err
		}
		return cmdutil.FlagErrorf("%w\nSeparate git clone flags with '--'.", err)
	})

	_ = cmd.MarkFlagRequired("owner")

	return cmd
}

func reposRun(opts *Options) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cmd, err := newCmd(opts)
	if err != nil {
		return err
	}

	return cmd.run(ctx)
}
