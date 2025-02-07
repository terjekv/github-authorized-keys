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

package cmd

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/terjekv/github-authorized-keys/config"
	"github.com/terjekv/github-authorized-keys/jobs"
	"github.com/terjekv/github-authorized-keys/server"
)

var cfgFile string

// ETCDTTLDefault - default ttl - 1day in seconds = 24 hours * 60 minutes * 60 seconds
const ETCDTTLDefault = int64(24 * 60 * 60)

// SyncUsersIntervalDefault - default interval between synchronize users - 5 minutes in seconds = 5 minutes * 60 seconds
const SyncUsersIntervalDefault = int64(5 * 60)

var flags = []flag{
	{"a", "string", "github_api_token", "", "Github API token       ( environment variable GITHUB_API_TOKEN could be used instead ) (read more https://github.com/blog/1509-personal-api-tokens)"},
	{"o", "string", "github_organization", "", "Github organization    ( environment variable GITHUB_ORGANIZATION could be used instead )"},
	{"n", "string", "github_admin_team_name", "", "Github admin team name ( environment variable GITHUB_ADMIN_TEAM_NAME could be used instead )"},
	{"N", "string", "github_user_team_name", "", "Github user team name  ( environment variable GITHUB_USER_TEAM_NAME could be used instead )"},
	{"i", "int", "github_admin_team_id", 0, "Github admin team id   ( environment variable GITHUB_ADMIN_TEAM_ID could be used instead )"},
	{"I", "int", "github_user_team_id", 0, "Github user team id    ( environment variable GITHUB_USER_TEAM_ID could be used instead )"},

	{"g", "strings", "sync_users_admin_groups", []string{}, "CSV groups name     ( environment variable SYNC_ADMIN_USERS_GROUPS could be used instead )"},
	{"G", "strings", "sync_users_users_groups", []string{}, "CSV groups name     ( environment variable SYNC_USERS_USERS_GROUPS could be used instead )"},

	{"s", "string", "sync_users_shell", "/bin/bash", "User shell 	    ( environment variable SYNC_USERS_SHELL could be used instead )"},
	{"r", "string", "sync_users_root", "/", "Root directory 	    ( environment variable SYNC_USERS_ROOT could be used instead )"},
	{"c", "int64", "sync_users_interval", SyncUsersIntervalDefault, "Sync each x sec     ( environment variable SYNC_USERS_INTERVAL could be used instead )"},

	{"e", "strings", "etcd_endpoint", []string{}, "CSV etcd endpoints  ( environment variable ETCD_ENDPOINT could be used instead )"},
	{"p", "string", "etcd_prefix", "/github-authorized-keys", "Path for etcd data  ( environment variable ETCD_PREFIX could be used instead )"},
	{"t", "int64", "etcd_ttl", ETCDTTLDefault, "ETCD value's ttl    ( environment variable ETCD_TTL could be used instead )"},

	{"d", "bool", "integrate_ssh", false, "Integrate with ssh  ( environment variable INTEGRATE_SSH could be used instead )"},
	{"l", "string", "listen", ":301", "Listen              ( environment variable LISTEN could be used instead )"},
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "github-authorized-keys",
	Short: "Use GitHub teams to manage system user accounts and authorized_keys",
	Long: `
Use GitHub teams to manage system user accounts and authorized_keys.

Config:
  REQUIRED: Github API token        | flag --github-api-token    OR environment variable GITHUB_API_TOKEN
  REQUIRED: Github organization     | flag --github-organization OR environment variable GITHUB_ORGANIZATION
  REQUIRED: One of
  		   Github admin team name | flag --github-admin-team-name OR environment variable GITHUB_ADMIN_TEAM_NAME
  			OR
  		   Github admin team id   | flag --github-admin-team-id OR Environment variable GITHUB_ADMIN_TEAM_ID
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.WithFields(log.Fields{"class": "RootCmd", "method": "RunE"})
		// @TODO Support viper duration type
		etcdTTL, err := time.ParseDuration(viper.GetString("etcd_ttl") + "s")

		if err != nil {
			return err
		}

		cfg := config.Config{
			GithubAPIToken:     viper.GetString("github_api_token"),
			GithubOrganization: viper.GetString("github_organization"),
			//			GithubTeamID:       viper.GetInt("github_team_id"),

			GithubAdminTeamName: viper.GetString("github_admin_team_name"),
			GithubUserTeamName:  viper.GetString("github_user_team_name"),
			GithubAdminTeamID:   viper.GetInt("github_admin_team_id"),
			GithubUserTeamID:    viper.GetInt("github_user_team_id"),

			EtcdEndpoints: fixStringSlice(viper.GetString("etcd_endpoint")),
			EtcdPrefix:    viper.GetString("etcd_prefix"),
			EtcdTTL:       etcdTTL,

			//			UserGID:    viper.GetString("sync_users_gid"),

			UserAdminGroups: fixStringSlice(viper.GetString("sync_users_admin_groups")),
			UserUserGroups:  fixStringSlice(viper.GetString("sync_users_users_groups")),

			UserShell: viper.GetString("sync_users_shell"),
			Root:      viper.GetString("sync_users_root"),
			Interval:  uint64(viper.GetInt64("sync_users_interval")),

			IntegrateWithSSH: viper.GetBool("integrate_ssh"),

			Listen: viper.GetString("listen"),
		}

		logger.Infof("Config: GithubAPIToken - %v", mask(cfg.GithubAPIToken))
		logger.Infof("Config: GithubOrganization - %v", mask(cfg.GithubOrganization))
		logger.Infof("Config: GithubAdminTeamName - %v", mask(cfg.GithubAdminTeamName))
		logger.Infof("Config: GithubUserTeamName - %v", mask(cfg.GithubUserTeamName))
		logger.Infof("Config: GithubAdminTeamID - %v", mask(fmt.Sprintf("%d", cfg.GithubAdminTeamID)))
		logger.Infof("Config: GithubUserTeamID - %v", mask(fmt.Sprintf("%d", cfg.GithubUserTeamID)))
		//		logger.Infof("Config: GithubTeamID - %v", mask(string(cfg.GithubTeamID)))
		logger.Infof("Config: EtcdEndpoints - %v", cfg.EtcdEndpoints)
		logger.Infof("Config: EtcdPrefix - %v", cfg.EtcdPrefix)
		logger.Infof("Config: EtcdTTL - %v seconds", cfg.EtcdTTL)
		//		logger.Infof("Config: UserGID - %v", cfg.UserGID)
		logger.Infof("Config: UserAdminGroups - %v", cfg.UserAdminGroups)
		logger.Infof("Config: UserUserGroups - %v", cfg.UserUserGroups)
		logger.Infof("Config: UserShell - %v", cfg.UserShell)
		logger.Infof("Config: Root - %v", cfg.Root)
		logger.Infof("Config: Interval - %v seconds", cfg.Interval)
		logger.Infof("Config: IntegrateWithSSH - %v", cfg.IntegrateWithSSH)
		logger.Infof("Config: Listen - %v", cfg.Listen)

		err = cfg.Validate()

		if err == nil {
			jobs.Run(cfg)
			server.Run(cfg)
		}

		return err
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Config file
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"Config file         (default is $HOME/.github-authorized-keys.yaml)")

	for _, f := range flags {
		createCmdFlags(RootCmd, f)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".github-authorized-keys") // name of config file (without extension)
	viper.AddConfigPath("$HOME")                   // adding home directory as first search path
	viper.AutomaticEnv()                           // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
