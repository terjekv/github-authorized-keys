#!/bin/sh

set -e

GAK_VERSION=0.0.25

[ -f /etc/os-release ] && source /etc/os-release

cd /tmp
CURL="curl -sSfOL"

# Fetch shared artifacts
$CURL "https://github.com/terjekv/github-authorized-keys/releases/download/v${GAK_VERSION}/github-authorized-keys-v${GAK_VERSION}-linux-amd64.tar.gz"
$CURL "https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/authorized-keys"

# Fetch systemd service file
$CURL "https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/github-authorized-keys.service"

if [ "${ID_LIKE}" == "fedora" ]; then
    echo "Installing on Fedora-like system, with systemd and SELinux."

    $CURL https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/env.rhel

    # SElinux
    $CURL https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/ssh_on_socket_301.pp
    sudo semodule -i ssh_on_socket_301.pp

    sudo mv env.rhel /etc/github-authorized-keys
else
    echo "Installing generic system, with systemd and no SELinux."

    $CURL -OL https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/env
    sudo mv env /etc/github-authorized-keys
fi

# Unpack binary
tar xvzf github-authorized-keys-v${GAK_VERSION}-linux-amd64.tar.gz


# Move artifacts into place
sudo mv github-authorized-keys /usr/local/sbin/github-authorized-keys
sudo mv authorized-keys /usr/local/sbin/authorized-keys
sudo mv github-authorized-keys.service /etc/systemd/system/github-authorized-keys.service

# Ensure permissions are correct
sudo chmod 755 /usr/local/sbin/authorized-keys
sudo chmod 755 /usr/local/sbin/github-authorized-keys

# Edit configuration, add token, organization, and team
sudo vi /etc/github-authorized-keys

sudo systemctl daemon-reload

if "x${ID_LIKE}" == "xfedora"; then
    echo "Edit /etc/github-authorized-keys to add your token, organization, and team".
else
    echo "Edit /etc/github-authorized-keys to configured your setup."
    echo "Ensure that token, organization, and team is set up, and that".
    echo "the templates for creating users and groups are correct for your system."
fi

echo "After that, run the following commands to enable and start the service".

echo sudo systemctl enable github-authorized-keys
echo sudo systemctl start github-authorized-keys