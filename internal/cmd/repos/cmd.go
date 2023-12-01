package repos

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/shurcooL/graphql"
	"golang.org/x/sync/errgroup"

	"github.com/zkhvan/gh-clone/internal/gh"
	"github.com/zkhvan/gh-clone/internal/widgets/progress"
)

type cmd struct {
	opts *Options
	gql  *api.GraphQLClient
}

func newCmd(opts *Options) (*cmd, error) {
	cmd := &cmd{opts: opts}

	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, err
	}
	cmd.gql = gql

	return cmd, nil
}

func (c *cmd) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	pg := progress.New()

	ch := make(chan string, 1)
	g.Go(func() error {
		return c.getRepos(ctx, ch, pg)
	})

	for i := 0; i < c.opts.Workers; i++ {
		g.Go(func() error {
			return c.cloneRepos(ctx, ch, pg)
		})
	}

	if err := pg.Wait(); err != nil {
		if errors.Is(err, progress.ErrAborted) {
			cancel()
			return nil
		}
		return err
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *cmd) getRepos(ctx context.Context, repositories chan<- string, pg *progress.Bar) error {
	defer close(repositories)

	var query struct {
		Search struct {
			RepositoryCount int
			PageInfo        struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []struct {
				Repository struct {
					NameWithOwner string
				} `graphql:"... on Repository"`
			}
		} `graphql:"search(type: REPOSITORY, query: $query, first: $first, after: $after)"`
	}

	variables := map[string]interface{}{
		"query": graphql.String(fmt.Sprintf("owner:%s archived:false", c.opts.Owner)),
		"first": graphql.Int(20),
		"after": (*graphql.String)(nil),
	}

	for {
		if err := c.gql.QueryWithContext(ctx, "GetRepositories", &query, variables); err != nil {
			return err
		}

		pg.SetTotal(query.Search.RepositoryCount)

		for _, n := range query.Search.Nodes {
			select {
			case repositories <- n.Repository.NameWithOwner:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if !query.Search.PageInfo.HasNextPage {
			break
		}

		variables["after"] = graphql.String(query.Search.PageInfo.EndCursor)
	}

	return nil
}

func (c *cmd) cloneRepos(ctx context.Context, repositories <-chan string, pg *progress.Bar) error {
	var err error
	for {
		select {
		case r, ok := <-repositories:
			if !ok {
				return err
			}

			pg.Start(r)

			err = errors.Join(err, c.cloneRepo(ctx, r))

			pg.Increment(r)
		case <-ctx.Done():
			err = errors.Join(err, ctx.Err())
			return err
		}
	}
}

func (c *cmd) cloneRepo(ctx context.Context, r string) error {
	repo, err := repository.Parse(r)
	if err != nil {
		return err
	}

	dir := filepath.Join(c.opts.Directory, repo.Owner, repo.Name)
	if c.exists(dir) {
		return c.updateRepo(ctx, repo, dir)
	}

	args := make([]string, 0, len(c.opts.GitArgs)+2)
	args = append(args, "repo", "clone", r, filepath.Join(c.opts.Directory, repo.Owner, repo.Name))
	if len(c.opts.GitArgs) > 0 {
		args = append(args, "--")
		args = append(args, c.opts.GitArgs...)
	}

	_, stderr, err := gh.ExecContext(ctx, args...)
	if err != nil {
		return fmt.Errorf("clone failed: %s: %w:\n%s", r, err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

func (c *cmd) exists(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}

	return true
}

func (c *cmd) updateRepo(ctx context.Context, repo repository.Repository, dir string) error {
	args := []string{"repo", "sync"}
	_, stderr, err := gh.ExecContextDirectory(ctx, dir, args...)
	if err != nil {
		return fmt.Errorf("sync failed: %s: %w:\n%s", dir, err, strings.TrimSpace(stderr.String()))
	}

	return nil
}
