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

package config

import (
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

// Config - structure to store global configuration
type Config struct {
	GithubAPIToken     string
	GithubOrganization string

	GithubAdminTeamName string
	GithubAdminTeamID   int
	GithubUserTeamName  string
	GithubUserTeamID    int

	EtcdEndpoints []string
	EtcdTTL       time.Duration
	EtcdPrefix    string

	UserAdminGroups []string
	UserUserGroups  []string

	UserShell string
	Root      string
	Interval  uint64

	IntegrateWithSSH bool

	Listen string
}

// Validate - process validation of config values
func (c Config) Validate() (err error) {
	err = validation.ValidateStruct(&c,
		validation.Field(&c.GithubAPIToken, validation.Required.Error("is required")),
		validation.Field(&c.GithubOrganization, validation.Required.Error("is required")))

	if err != nil {
		return
	}

	// Validate Github Team exists
	if c.GithubAdminTeamName == "" && c.GithubUserTeamName == "" {
		err = errors.New("either a github admin team name or a github user team name is required")
	}
	return
}
