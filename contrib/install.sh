#!/bin/sh

set -e

GAK_VERSION=0.0.25
BRANCH="${$1:-'main'}"
# terjekv/issue-35-Ease-of-installation

[ -f /etc/os-release ] && source /etc/os-release

cd /tmp
CURL="curl -sSfOL"
BINARY_PATH="/usr/local/sbin"
# Loading a given configuration file from --config doesn't work, so we use the hardcoded default
# path for the configuration file...
CONF_FILE="/root/.github-authorized-keys.yaml" 
SELINUX_POLICIES="my-curl.pp allow-github-authorized-keys-from-usr-local.pp"

# https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/
RAW_CONTRIB_URL=https://raw.githubusercontent.com/terjekv/github-authorized-keys/${BRANCH}/contrib

# Fetch shared artifacts
$CURL "https://github.com/terjekv/github-authorized-keys/releases/download/v${GAK_VERSION}/github-authorized-keys-v${GAK_VERSION}-linux-amd64.tar.gz"
$CURL "${RAW_CONTRIB_URL}/authorized-keys"

# Fetch systemd service file
$CURL "${RAW_CONTRIB_URL}/github-authorized-keys.service"

if [ "${ID_LIKE}" == "fedora" ]; then
    echo "** Detected Fedora-like system, with systemd and SELinux."

    echo "  - Configuring SELinux policies"
    $CURL "${RAW_CONTRIB_URL}/env.rhel"

    # SElinux
    for policy in $SELINUX_POLICIES; do
        $CURL "${RAW_CONTRIB_URL}/${policy}"
        sudo semodule -i ${policy}
        rm ${policy}
    done

    $CURL "${RAW_CONTRIB_URL}/ssh_on_socket_301.pp"
    $CURL "${RAW_CONTRIB_URL}/allow-github-authorized-keys-from-usr-local.pp"
    sudo semodule -i ssh_on_socket_301.pp
    sudo semodule -i allow-github-authorized-keys-from-usr-local.pp

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


if [ -f ${CONF_FILE} ] && ! grep 'GITHUB_API_TOKEN: ghp_token' ${CONF_FILE} > /dev/null; then
    echo
    echo "Warning: Skipping installation of default environment file, as it is already modified."
    echo "If you want to use the default environment file, please remove ${CONF_FILE} and run this script again."
    echo 
    ENV_ALREADY_SET=True
else
    echo "  - Installing environment file."
    # YAMLIFY the environment file
    echo "---" > ${ENV_FILE}.yaml
    awk '{sub(/=/,": ");}1' < ${ENV_FILE} >> ${ENV_FILE}.yaml
    sudo mv ${ENV_FILE}.yaml ${CONF_FILE}
    rm ${ENV_FILE}
fi

# Ensure permissions are correct
sudo chmod 755 ${BINARY_PATH}/authorized-keys
sudo chmod 755 ${BINARY_PATH}/github-authorized-keys

sudo systemctl daemon-reload

if [ "${ID_LIKE}" == "fedora" ]; then
    echo "  - Ensuring SELinux permissions are correct"
    # Ensure SELinux permissions are correct
    sudo /sbin/restorecon -v ${BINARY_PATH}/github-authorized-keys
fi

echo
echo "** Installation complete! **"
echo

if ! [ "${ENV_ALREADY_SET}" == "True" ]; then
    echo "You need to configure the service before it can be used!"
    if [ "${ID_LIKE}" == "fedora" ]; then
        echo "Edit ${CONF_FILE} to add your token, organization, and team".    
    else
        echo "Edit ${CONF_FILE} to configured your setup."
        echo "Ensure that token, organization, and team is set up, and that".
        echo "the templates for creating users and groups are correct for your system."
    fi
    echo "After that run the following commands to enable and start the service".

    echo sudo systemctl enable github-authorized-keys
    echo sudo systemctl start github-authorized-keys
else
    echo "Configuration is already set up, enabling and starting the service."
#    sudo systemctl enable github-authorized-keys
#    sudo systemctl start github-authorized-keys
fi

