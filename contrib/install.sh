#!/bin/sh

set -e

GAK_VERSION=0.0.25

[ -f /etc/os-release ] && source /etc/os-release

cd /tmp
CURL="curl -sSfOL"
BINARY_PATH="/usr/local/sbin"

# https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/
RAW_CONTRIB_URL=https://raw.githubusercontent.com/terjekv/github-authorized-keys/terjekv/issue-35-Ease-of-installation/contrib

# Fetch shared artifacts
$CURL "https://github.com/terjekv/github-authorized-keys/releases/download/v${GAK_VERSION}/github-authorized-keys-v${GAK_VERSION}-linux-amd64.tar.gz"
$CURL "${RAW_CONTRIB_URL}/authorized-keys"

# Fetch systemd service file
$CURL "${RAW_CONTRIB_URL}/github-authorized-keys.service"

if [ "${ID_LIKE}" == "fedora" ]; then
    echo "** Detected Fedora-like system, with systemd and SELinux."

    $CURL "${RAW_CONTRIB_URL}/env.rhel"

    # SElinux
    $CURL "${RAW_CONTRIB_URL}/ssh_on_socket_301.pp"
    sudo semodule -i ssh_on_socket_301.pp

    ENV_FILE=env.rhel
else
    echo "** Detected generic system, with systemd and no SELinux."

    $CURL -OL "${RAW_CONTRIB_URL}/env"
    ENV_FILE=env
fi

# Unpack binary
tar xzf github-authorized-keys-v${GAK_VERSION}-linux-amd64.tar.gz

echo "  - Installing binary files into ${BINARY_PATH}"
# Move artifacts into place
sudo mv github-authorized-keys ${BINARY_PATH}/github-authorized-keys

sudo mv authorized-keys ${BINARY_PATH}/authorized-keys

echo "  - Installing systemd service."
sudo mv github-authorized-keys.service /etc/systemd/system/github-authorized-keys.service


if ! grep 'GITHUB_API_TOKEN=ghp_token' /etc/github-authorized-keys > /dev/null; then
    echo
    echo "Warning: Skipping installation of default environment file, as it is already modified."
    echo "If you want to use the default environment file, please remove /etc/github-authorized-keys and run this script again."
    echo 
else
    echo "  - Installing environment file."
    sudo mv ${ENV_FILE} /etc/github-authorized-keys
fi

# Ensure permissions are correct
sudo chmod 755 ${BINARY_PATH}/authorized-keys
sudo chmod 755 ${BINARY_PATH}/github-authorized-keys

sudo systemctl daemon-reload

echo "** Installation complete! **"
echo
echo "You need to configure the service before it can be used!"
if [ "${ID_LIKE}" == "fedora" ]; then
    echo "Edit /etc/github-authorized-keys to add your token, organization, and team".
    
else
    echo "Edit /etc/github-authorized-keys to configured your setup."
    echo "Ensure that token, organization, and team is set up, and that".
    echo "the templates for creating users and groups are correct for your system."
fi

echo
echo "After that, run the following commands to enable and start the service".

echo sudo systemctl enable github-authorized-keys
echo sudo systemctl start github-authorized-keys