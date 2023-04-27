# Contrib

Bits and pieces that are helpful for getting github-authorized-keys up and running.

## Install.sh

A script that installs github-authorized-keys to /usr/local/sbin and sets up a systemd service file. It also installs the SELinux policy file and sets up file contexts if SELinux is detected.

For most installations, this is very simple solution to getting github-authorized-keys up and running.

Note that SELinux support for this script is limited to `SE Linux modular policy version 1, mod version 19`. This should be fine for RHEL 8 and newer derivatives, but may not work on older systems (RHEL 7 and friends). If you are running an older system, you can still use the script, but you will have to modify it to use your own policy file (see below).

A typical way to call the script is to do:

```bash
GAK_GITHUB_API_TOKEN="ghp_..." GAK_GITHUB_ORGANIZATION="myorgname" GAK_GITHUB_ADMIN_TEAM_NAME="myadminteam" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/install.sh)"
```

If you want to give the values interactively, you can do:

```bash
GAK_INTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/install.sh)"
```

You may also combine the two:

```bash
GAK_INTERACTIVE=1 GAK_GITHUB_API_TOKEN="ghp_..." GAK_GITHUB_ORGANIZATION="myorgname" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/install.sh)"
```

When issuing interactive usage, the defaults the script provides will be the same as it would have chosen if the script was run non-interactively.
The order of precedence for values are as follows:

1. Existing values from an already existing configuration file
2. Environment variables mapped to variables in the script, prefixed with `GAK_`
3. Values from the environment file selected for the platform (rhel/debian)

Any value issued interactively will be selected above the defaults.


## Environment files

Two examples, [`env.rhel`](env.rhel) and [`env`](env) (debian-like). These can be used to set up
the environment for `github-authorized-keys`. If you are planning to use a YAML configuration
file, you can convert the environment variables to YAML as follows:

```bash
awk '{sub(/=/,": ");}1' < ${ENV_FILE} >> ${ENV_FILE}.yaml
```

## Systemd

`github-authorized-keys.service` is a systemd service file that can be used either as-is
or as a template for your own service file. It assumes that github-authorized-keys is
run from /usr/local/sbin, which on some systems requires a custom SELinux policy (see below).

## SELinux

### File contexts

The following file contexts are required for github-authorized-keys to run:

- `/usr/local/sbin/github-authorized-keys`: `bin_t`
- `/usr/local/sbin/authorized-keys`: `bin_t`

`install.sh` will install these file contexts if SELinux is detected, by running `semanage fcontext` and `restorecon`. Existing contexts for these two files will be replaced.

### Policies

There is a default policy file available:

- [github-authorized-keys-allow-sshd-reserved-ports.pp](github-authorized-keys-allow-sshd-reserved-ports.pp):
  This policy allows curl access to reserved ports when run by sshd.

- [github-authorized-keys-allow-sshd-reserved-ports.te](github-authorized-keys-allow-sshd-reserved-ports.te):
  This is the type enforcement file that can be used to compile the policy file. This is useful if you want
  to ensure that you are installing a policy file that is compatible with your system. However, you will have to
  modify install.sh to use your self-complied policy file.


