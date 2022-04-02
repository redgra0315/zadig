package gitee

import (
	"context"

	giteeClient "gitee.com/openeuler/go-gitee/gitee"
	"github.com/antihax/optional"
	"go.uber.org/zap"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/code/client"
	"github.com/koderover/zadig/pkg/tool/gitee"
)

type Config struct {
	AccessToken string `json:"access_token"`
	EnableProxy bool   `json:"enable_proxy"`
}

type Client struct {
	Client      *gitee.Client
	AccessToken string
}

func (c *Config) Open(id int, logger *zap.SugaredLogger) (client.CodeHostClient, error) {
	client := gitee.NewClient(id, c.AccessToken, config.ProxyHTTPSAddr(), c.EnableProxy)
	return &Client{
		Client:      client,
		AccessToken: c.AccessToken,
	}, nil
}

func (c *Client) ListBranches(opt client.ListOpt) ([]*client.Branch, error) {
	bList, err := c.Client.ListBranches(context.TODO(), opt.Namespace, opt.ProjectName, nil)
	if err != nil {
		return nil, err
	}
	var res []*client.Branch
	for _, o := range bList {
		res = append(res, &client.Branch{
			Name:      o.Name,
			Protected: o.Protected,
		})
	}
	return res, nil
}

func (c *Client) ListTags(opt client.ListOpt) ([]*client.Tag, error) {
	tags, err := c.Client.ListTags(context.TODO(), c.AccessToken, opt.Namespace, opt.ProjectName)
	if err != nil {
		return nil, err
	}
	var resp []*client.Tag
	for _, tag := range tags {
		resp = append(resp, &client.Tag{
			Name:    tag.Name,
			Message: tag.Message,
		})
	}

	return resp, nil
}

func (c *Client) ListPrs(opt client.ListOpt) ([]*client.PullRequest, error) {
	prs, err := c.Client.ListPullRequests(context.TODO(), opt.Namespace, opt.ProjectName, &giteeClient.GetV5ReposOwnerRepoPullsOpts{
		PerPage: optional.NewInt32(100),
	})
	if err != nil {
		return nil, err
	}
	var res []*client.PullRequest
	for _, o := range prs {
		res = append(res, &client.PullRequest{
			ID:             int(o.Number),
			State:          o.State,
			User:           o.User.Login,
			Number:         int(o.Number),
			AuthorUsername: o.User.Login,
			Title:          o.Title,
			SourceBranch:   o.Base.Ref,
			TargetBranch:   o.Base.Ref,
		})
	}
	return res, nil
}

func (c *Client) ListNamespaces(keyword string) ([]*client.Namespace, error) {
	user, err := c.Client.GetAuthenticatedUser(context.TODO())
	if err != nil {
		return nil, err
	}
	namespaceUser := client.Namespace{
		Name: user.Login,
		Path: user.Login,
		Kind: client.UserKind,
	}

	organizations, err := c.Client.ListOrganizationsForAuthenticatedUser(context.TODO())
	if err != nil {
		return nil, err
	}

	var res []*client.Namespace
	res = append(res, &namespaceUser)
	for _, o := range organizations {
		res = append(res, &client.Namespace{
			Name: o.Login,
			Path: o.Login,
			Kind: client.OrgKind,
		})
	}
	return res, nil
}

func (c *Client) ListProjects(opt client.ListOpt) ([]*client.Project, error) {
	projects, err := c.Client.ListRepositoriesForAuthenticatedUser(c.AccessToken, opt.Key, opt.Page, opt.PerPage)
	if err != nil {
		return nil, err
	}
	var res []*client.Project
	for _, project := range projects {
		res = append(res, &client.Project{
			ID:            project.ID,
			Name:          project.Name,
			DefaultBranch: project.DefaultBranch,
		})
	}

	return res, nil
}
