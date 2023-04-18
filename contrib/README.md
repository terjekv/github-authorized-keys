# Contrib

Bits and pieces that are helpful for getting github-authorized-keys up and running.

## Install.sh

A simple script that installs github-authorized-keys to /usr/local/sbin and sets up a
systemd service file. It also installs the SELinux policy files if SELinux is detected.
For most installations, this is very simple solution to getting github-authorized-keys
up and running.



## Environment files

Two examples, [`env.rhel`](env.rhel) and [`env`](env) (generic, mostly debian-like). These can be used to set up
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

There are two policy files available:

- [my-curl.pp](my-curl.pp): This policy allows curl to access the github API when called as
  a part of sshd.

- [allow-github-authorized-keys-from-usr-local.pp](allow-github-authorized-keys-from-usr-local.pp):
  This policy allows github-authorized-keys to be executed by systemd while located in /usr/local/sbin.

For default installations on SELinux systems, the first is required. You can avoid the latter if you
install github-authorized-keys to /usr/sbin instead of /usr/local/sbin. However, this will require 
a number of manual steps to get the service file set up correctly.