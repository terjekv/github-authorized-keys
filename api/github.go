/*
 * Github Authorized Keys - Use GitHub teams to manage system user accounts and authorized_keys
 *
 * Copyright 2016 Cloud Posse, LLC <hello@cloudposse.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"context"
	"errors"
	"strings"

	"github.com/google/go-github/v43/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var (
	// ErrorGitHubConnectionFailed - returned when there was a connection error with github.com
	ErrorGitHubConnectionFailed = errors.New("Connection to github.com failed")

	// ErrorGitHubAccessDenied - returned when there was access denied to github.com resource
	ErrorGitHubAccessDenied = errors.New("Access denied")

	// ErrorGitHubNotFound - returned when github.com resource not found
	ErrorGitHubNotFound = errors.New("Not found")
)

func init() {
	viper.SetDefault("github_api_max_page_size", 100)
}

// Naive oauth setup
func newAccessToken(token string) oauth2.TokenSource {
	return oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
}

// GithubClient - client for operate with Github API
type GithubClient struct {
	client         *github.Client
	owner          string
	organizationId *int64
}

// GetTeam - return team structure based on name or id
func (c *GithubClient) GetTeam(name string, id int) (team *github.Team, err error) {
	defer func() {
		if r := recover(); r != nil {
			team = nil
			err = ErrorGitHubConnectionFailed
		}
	}()

	team = nil
	err = nil

	var opt = &github.ListOptions{
		PerPage: viper.GetInt("github_api_max_page_size"),
	}

	for {
		teams, response, _ := c.client.Teams.ListTeams(context.Background(), c.owner, opt)

		if response.StatusCode != 200 {
			err = ErrorGitHubAccessDenied
			return
		}

		for _, localTeam := range teams {
			if *localTeam.ID == int64(id) || *localTeam.Slug == name {
				team = localTeam
				// team found
				return
			}
		}

		if response.LastPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}
	// Exit with error

	err = errors.New("No such team name or id could be found")
	return
}

func (c *GithubClient) getUser(name string) (*github.User, error) {
	user, response, err := c.client.Users.Get(context.Background(), name)

	if response.StatusCode != 200 {
		return nil, ErrorGitHubAccessDenied
	}

	return user, err
}

// IsTeamMember - check if {user} is a member of {team}
func (c *GithubClient) IsTeamMember(user string, team *github.Team) (bool, error) {
	result, _, err := c.client.Teams.GetTeamMembershipByID(
		context.Background(), *c.organizationId, *team.ID, user,
	)
	if result != nil {
		return true, err
	}

	if err != nil && strings.Contains(err.Error(), "Not Found") {
		return false, nil
	}

	return false, err
}

// GetKeys - return array of user's {userName} public keys
func (c *GithubClient) GetKeys(userName string) (keys []*github.Key, err error) {
	defer func() {
		if r := recover(); r != nil {
			keys = make([]*github.Key, 0)
			err = ErrorGitHubConnectionFailed
		}
	}()

	logger := log.WithFields(log.Fields{"class": "GithubClient", "method": "Get"})

	var opt = &github.ListOptions{
		PerPage: viper.GetInt("github_api_max_page_size"),
	}

	for {
		items, response, localErr := c.client.Users.ListKeys(context.Background(), userName, opt)

		logger.Debugf("Response: %v", response)
		logger.Debugf("Response.StatusCode: %v", response.StatusCode)

		switch response.StatusCode {
		case 200:
			keys = append(keys, items...)
		case 404:
			err = ErrorGitHubNotFound
			return
		default:
			err = ErrorGitHubAccessDenied
			return
		}

		if localErr != nil {
			err = localErr
			return
		}

		if response.LastPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}

	return
}

// GetTeamMembers - return array of user's that are {team} members
func (c *GithubClient) GetTeamMembers(team *github.Team) (users []*github.User, err error) {
	defer func() {
		if r := recover(); r != nil {
			users = make([]*github.User, 0)
			err = ErrorGitHubConnectionFailed
		}
	}()

	var opt = &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{
			PerPage: viper.GetInt("github_api_max_page_size"),
		},
	}

	for {
		members, resp, localErr := c.client.Teams.ListTeamMembersByID(
			context.Background(), *c.organizationId, *team.ID, opt,
		)
		if resp.StatusCode != 200 {
			return nil, ErrorGitHubAccessDenied
		}
		if localErr != nil {
			err = localErr
			return
		}

		users = append(users, members...)

		if resp.LastPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return
}

func (client *GithubClient) SetOrganizationID() error {
	if client.organizationId != nil {
		return nil
	}

	organization, response, err := client.client.Organizations.Get(context.Background(), client.owner)
	if response != nil && response.StatusCode != 200 {
		return ErrorGitHubAccessDenied
	}
	if err != nil {
		return err
	}

	client.organizationId = organization.ID

	return err
}

// NewGithubClient - constructor of GithubClient structure
func NewGithubClient(token, owner string) *GithubClient {
	c := oauth2.NewClient(context.Background(), newAccessToken(token))
	client := &GithubClient{client: github.NewClient(c), owner: owner}
	err := client.SetOrganizationID()
	if err != nil {
		log.Info("GitHub Client initialization failed. Error: ", err)
	}

	return client
}
