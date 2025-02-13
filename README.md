# Github Authorized Keys [![Build Status](https://github.com/terjekv/github-authorized-keys/actions/workflows/build.yml/badge.svg)](https://github.com/terjekv/github-authorized-keys/)

Use GitHub teams to manage system user accounts and `authorized_keys`. 


[![Go Report Card](https://goreportcard.com/badge/github.com/terjekv/github-authorized-keys)](https://goreportcard.com/report/github.com/terjekv/github-authorized-keys)
[![Docker Pulls](https://img.shields.io/docker/pulls/terjekv/github-authorized-keys.svg)](https://hub.docker.com/r/terjekv/github-authorized-keys)
[![GitHub Stars](https://img.shields.io/github/stars/terjekv/github-authorized-keys.svg)](https://github.com/terjekv/github-authorized-keys/stargazers) 
[![GitHub Issues](https://img.shields.io/github/issues/terjekv/github-authorized-keys.svg)](https://github.com/terjekv/github-authorized-keys/issues)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://github.com/terjekv/github-authorized-keys/pulls)
[![License](https://img.shields.io/badge/license-APACHE%202.0%20-brightgreen.svg)](https://github.com/terjekv/github-authorized-keys/blob/master/LICENSE)
<!-- [![Coverage Status](https://coveralls.io/repos/github/terjekv/github-authorized-keys/badge.svg?branch=main)](https://coveralls.io/github/terjekv/github-authorized-keys?branch=main) -->

----

## Screenshots

**Administrators** 
* Automatically provision new users to production servers simply by adding them to a designated GitHub team (e.g. `ssh`). 
  ![Demo](docs/github-team-demo.png)
* No need to keep `authorized_keys` up to date because keys are pulled directly from github.com API and *optionally* cached in etcd
* Immediately revoke SSH access to servers by evicting users from the GitHub team
* Easy to deploy


**End Users**
* Self-administer public SSH keys via the [GitHub account settings](https://github.com/settings/keys).
  ![Demo](docs/github-keys-demo.png)
* No need to manage multiple SSH keys


## Architecture

This tool consists of three parts:

1. User Account / Authorized Keys provisioner which polls [GitHub API for users](https://developer.github.com/v3/users/keys/) that correspond to a given GitHub Organization & Team using a [personal access token](https://github.com/settings/tokens). It's responsible for adding or removing users from the system. All commands are templatized to allow it to run on multiple distributions. 
2. Simple read-only REST API that provides public keys for users, which is used by the `AuthorizedKeysCommand` in the `sshd_config`; this allows you to expose the service internally without compromising your Github Token. The public SSH access keys are *optionally* cached in Etcd for performance and reliability.
3. An `AuthorizedKeysCommand` [script](contrib/authorized-keys) that will `curl` the REST API for a user's public keys.

## Getting Started

### Direct installation

If you are running a derivative of RHEL9 (Fedora 36+, CentOS 9*, Rocky 9*, etc) or modern Debian derivatives (Ubuntu etc)
the install script should install and configure the service for you.

```bash
$ sudo /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/install.sh)"
```

After the install is finished, you will be prompted to edit the configuration file `/root/.github-authorized-keys.yaml` and
add your GitHub API token, organization name, and team name.

For older versions of RHEL/CentOS, you will need to adapt the SELinux policies. Work is being done to make this easier.

Note: On other distributions, you will also need to validate the templates for adding users `LINUX_USER_ADD_TPL`, adding users with a
GID available `LINUX_USER_ADD_WITH_GID_TPL` and adding users to groups `LINUX_USER_ADD_TO_GROUP_TPL` are correct for your
distribution. 


### Docker

An easy way to get up and running is by using the ready-made docker container. The only dependency is [Docker](https://docs.docker.com/engine/installation) itself. If you wish to run [CoreOS](docs/coreos.md) or use `systemd`, there's a [sample unit file](contrib/github-authorized-keys-docker.service).

A prebuilt public [docker image](https://hub.docker.com/r/terjekv/github-authorized-keys/) that is built using upon tagging a release (via [releases.yml](.github/workflows/releases.yml)) or you can build your own from source.

```
docker build -t terjekv/github-authorized-keys .
```

### Running GitHub Authorized Keys

All arguments can be passed both as environment variables or command-line arguments, or even mix-and-match them to suit your tastes.

Available configuration options:

| **Environment Variable**  | **Argument**                | **Description**                                  | **Default**              |
| ------------------------- | --------------------------- | ------------------------------------------------ | ------------------------ |
| `GITHUB_API_TOKEN`        | `--github-api-token`        | GitHub API Token (read-only)                     |                          |
| `GITHUB_ORGANIZATION`     | `--github-organization`     | GitHub Organization Containing Team              |                          |
| `GITHUB_ADMIN_TEAM_NAME`  | `--github-admin-team-name`  | Name of GitHub Team that grants admin SSH access |                          |
| `GITHUB_USER_TEAM_NAME`   | `--github-user-team-name`   | Name of GitHub Team that grants user SSH access  |                          |
| `GITHUB_ADMIN_TEAM_ID`    | `--github-admin-team-id`    | ID of GitHub Team that grants admin SSH access   |                          |
| `GITHUB_USER_TEAM_ID`     | `--github-user-team-id`     | ID of Github Team that grants user SSH access    |                          |
| `SYNC_USERS_ADMIN_GROUPS` | `--sync-users-admin-groups` | Default groups for admins                        | `wheel`                  |
| `SYNC_USERS_USERS_GROUPS` | `--sync-users-users-groups` | Default groups for users                         | `users`                  |
| `SYNC_USERS_SHELL`        | `--sync-users-shell`        | Default Login Shell                              | `/bin/bash`              |
| `SYNC_USERS_ROOT`         | `--sync-users-root`         | `chroot` path for user commands                  | `/`                      |
| `SYNC_USERS_INTERVAL`     | `--sync-users-interval`     | Interval used to update user accounts            | `300`                    |
| `ETCD_ENDPOINT`           | `--etcd-endpoint`           | Etcd endpoint used for caching public keys       |                          |
| `ETCD_TTL`                | `--etcd-ttl`                | Duration (in seconds) to cache public keys       | `86400`                  |
| `ETCD_PREFIX`             | `--etcd-prefix`             | Prefix for public keys stored in etcd            | `github-authorized-keys` |
| `LISTEN`                  | `--listen`                  | Bind address used for REST API                   | `:301`                   |
| `INTEGRATE_SSH`           | `--integrate-ssh`           | Flag to automatically configure SSH              | `false`                  |
| `LOG_LEVEL`               | `--log-level`               | Ccontrol the logging verbosity.                  | `info`                   |

## Quick Start 

We recommend that you specify all parameters as environment variables. If using `docker`, pass the [environment file](contrib/env) to the container using the `--env-file` argument.

Obtain the GitHub API Token (aka Personal Access Token) [here](https://github.com/settings/tokens). Click "Generate new token" and select `read:org`. That's it!

![Personal Access Token Permissions](docs/personal-access-token.png)


For example, [`/etc/github-authorized-keys`](contrib/env), might look like this:

```
GITHUB_API_TOKEN={token}
GITHUB_ORGANIZATION={organization}
GITHUB_ADMIN_TEAM_NAME=ssh
GITHUB_USER_TEAM_NAME=users
SYNC_USERS_ADMIN_GID=500
SYNC_USERS_ADMIN_GROUPS=sudo
SYNC_USERS_SHELL=/bin/bash
SYNC_USERS_ROOT=/host
SYNC_USERS_INTERVAL=300
ETCD_ENDPOINT=http://localhost:2739
ETCD_TTL=86400
ETCD_PREFIX=github-authorized-keys
LISTEN=:301
INTEGRATE_SSH=true
```

Then you could start it like this:

```
docker run \
  --volume /:/host \
  --expose "127.0.0.1:301:301" \
  --env-file /etc/github-authorized-keys \
     terjekv/github-authorized-keys:latest
```

**IMPORTANT** Remember to expose the REST API so you can retrieve user's public keys. Only public keys belonging to users found in the GitHub team will be returned.

**Note:** depending on your OS distribution, you might need to tweak the command templates. Keep reading for details.

## SELinux

On platforms running SELinux, you will need to add a policy to allow SSH on port 301 (if using the default port).
You can use the policy file `contrib/ssh_on_socket_301.pp` by adding it thusly: `semodule -i ssh_on_socket_301.pp`.

## Usage Examples

### Automatically Configure SSH

To leverage the `github-authorized-keys` API, we need to make a small tweak to the `sshd_config`. 

This can be done automatically by passing the `--integrate-ssh` flag (or setting `INTEGRATE_SSH=true`)

After modifying the `sshd_config`, it's necessary to restart the SSH daemon. This happens automatically by calling the `SSH_RESTART_TPL` command. Since this differs depending on the OS distribution, you can change the default behavior by setting the `SSH_RESTART_TPL` environment variable (default: `/usr/sbin/service ssh force-reload`). Similarly, you might need to tweak the `AUTHORIZED_KEYS_COMMAND_TPL` environment variable to something compatible with your OS.


### Manually Configure SSH

If you wish to manually configure your `sshd_config`, here's all you need to do:

```
AuthorizedKeysCommand /usr/local/sbin/authorized-keys
AuthorizedKeysCommandUser root
```

Then install a [wrapper script](contrib/authorized-keys) to `/usr/local/sbin/authorized-keys`. 

**Note**: this command requires `curl` to access the REST API in order to fetch authorized keys

### Etcd Fallback Cache

The REST API supports Etcd as cache for public keys. This mitigates any connectivity problems with GitHub's API. By default, the caching is disabled.

### Command Templates

Due to the vast differences between OS commands, the defaults provided might not work for you flavor of Linux.

Below are some of the settings which can be tweaked. 

| Environment Variable          | **Description**                                                                 | **Default**                                                                        |
| ----------------------------- | ------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `LINUX_USER_ADD_TPL`          | Command used to add a user to the system when no default group supplied.        | `adduser {username} --disabled-password --force-badname --shell {shell}`           |
| `LINUX_USER_ADD_WITH_GID_TPL` | Command used to add a user to the system when a default primary gid supplied  . | `adduser {username} --disabled-password --force-badname --shell {shell} --gid {gid | group}` |
| `LINUX_USER_ADD_TO_GROUP_TPL` | Command used to add the user to secondary groups                                | `adduser {username} {group}`                                                       |
| `LINUX_USER_DEL_TPL`          | Command used to delete a user from the system when removed the the team         | `deluser {username}`                                                               |
| `SSH_RESTART_TPL`             | Command used to restart SSH when `INTEGRATE_SSH=true`                           | `/usr/sbin/service ssh force-reload`                                               |
| `AUTHORIZED_KEYS_COMMAND_TPL` | Command used to fetch a user's `authorized_keys` from REST API                  | `/usr/bin/github-authorized-keys`                                                  |

The values in `{braces}` are macros that will be automatically substituted at run-time.

| **Macro**    | **Description**           |
| ------------ | ------------------------- |
| `{username}` | User's login name         |
| `{shell}`    | User's login shell        |
| `{group}`    | User's primary group name |
| `{gid}`      | User's primary group id   |

## Help

**Got a question?** 

File a GitHub [issue](https://github.com/terjekv/github-authorized-keys/issues). :)

## Contributing

### Bug Reports & Feature Requests

Please use the [issue tracker](https://github.com/terjekv/github-authorized-keys/issues) to report any bugs or file feature requests.

### Developing

The original use case from CloudPosse has for them been replaced by Teleport (https://github.com/cloudposse/github-authorized-keys/issues/34). This repository is maintained for personal use for time being.

In general, PRs are welcome.

 1. **Fork** the repo on GitHub
 2. **Clone** the project to your own machine
 3. **Commit** changes to your own branch
 4. **Push** your work back up to your fork
 5. Submit a **Pull request** so that changes can be reviewed.

**NOTE:** Be sure to merge the latest from "upstream" before making a pull request!

Here's how to get started...

1. `git clone https://github.com/terjekv/github-authorized-keys.git` to pull down the repository 
2. Review the [documentation](docs/) on compiling.

## License

[APACHE 2.0](LICENSE) © 2016-2017 [Cloud Posse, LLC](https://cloudposse.com)

    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at
     
      http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.

## About

GitHub Authorized Keys was originally maintained and funded by [Cloud Posse, LLC][website]. Like it? Please let them know at <hello@cloudposse.com>

We love [Open Source Software](https://github.com/cloudposse/)! 

See [our other projects][community] or [hire us][hire] to help build your next cloud-platform.

  [website]: http://cloudposse.com/
  [community]: https://github.com/cloudposse/
  [hire]: http://cloudposse.com/contact/
  
### Contributors


| [![Erik Osterman][erik_img]][erik_web]<br/>[Erik Osterman][erik_web] | [![Igor Rodionov][igor_img]][igor_web]<br/>[Igor Rodionov][igor_web] |
| -------------------------------------------------------------------- | -------------------------------------------------------------------- |

  [erik_img]: http://s.gravatar.com/avatar/88c480d4f73b813904e00a5695a454cb?s=144
  [erik_web]: https://github.com/osterman/
  [igor_img]: http://s.gravatar.com/avatar/bc70834d32ed4517568a1feb0b9be7e2?s=144
  [igor_web]: https://github.com/goruha/


