#!/bin/sh

set -e

if [ "$(id -u)" -ne 0 ]; then
        echo 'This script must be run by root' >&2
        exit 1
fi

GAK_VERSION=0.0.25
BRANCH="${1:-main}"

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

echo
echo "Installing github-authorized-keys v${GAK_VERSION}..."
echo
echo "  - Detected OS: ${PRETTY_NAME}"

if [ "${ID_LIKE}" == "fedora" ]; then
    echo "  - System is Fedora-like, expecting systemd and SELinux."
    echo "  - Configuring SELinux policies."
    $CURL "${RAW_CONTRIB_URL}/env.rhel"

    # SElinux
    for policy in $SELINUX_POLICIES; do
        $CURL "${RAW_CONTRIB_URL}/${policy}"
        semodule -i ${policy}
        rm ${policy}
    done

    ENV_FILE=env.rhel
else
    echo "  - No specific system recognized, assuming systemd but no SELinux."

    $CURL -OL "${RAW_CONTRIB_URL}/env"
    ENV_FILE=env
fi

# Unpack binary
tar xzf github-authorized-keys-v${GAK_VERSION}-linux-amd64.tar.gz
rm github-authorized-keys-v${GAK_VERSION}-linux-amd64.tar.gz README.md LICENSE

echo "  - Installing binary files into ${BINARY_PATH}."
# Move artifacts into place
mv github-authorized-keys ${BINARY_PATH}/github-authorized-keys

mv authorized-keys ${BINARY_PATH}/authorized-keys

echo "  - Installing systemd service."
mv github-authorized-keys.service /etc/systemd/system/github-authorized-keys.service


if test -f ${CONF_FILE} && ! grep -Fq 'ghp_token' ${CONF_FILE}; then
    echo "  - NOTE: ${CONF_FILE} already has configuration, staying clear."
    ENV_ALREADY_SET=True
else
    echo "  - Installing a fresh configuration into ${CONF_FILE}."
    # YAMLIFY the environment file
    echo "---" > ${ENV_FILE}.yaml
    awk '{sub(/=/,": ");}1' < ${ENV_FILE} >> ${ENV_FILE}.yaml
    mv ${ENV_FILE}.yaml ${CONF_FILE}
    rm ${ENV_FILE}
fi

# Ensure permissions are correct
chmod 755 ${BINARY_PATH}/authorized-keys
chmod 755 ${BINARY_PATH}/github-authorized-keys

systemctl daemon-reload

if [ "${ID_LIKE}" == "fedora" ]; then
    echo "  - Ensuring SELinux permissions are correct."
    /sbin/restorecon -v ${BINARY_PATH}/github-authorized-keys > /dev/null
    /sbin/restorecon -v ${BINARY_PATH}/authorized-keys > /dev/null
fi

echo "  - Validating ssh configuration."

if ! grep -Eq '^AuthorizedKeysCommand' /etc/ssh/sshd_config; then
    echo "    - Adding AuthorizedKeysCommand to sshd_config."
    echo "AuthorizedKeysCommand ${BINARY_PATH}/authorized-keys" >> /etc/ssh/sshd_config 
else
    echo "    - AuthorizedKeysCommand already set up, skipping."
fi

if ! grep -Eq '^AuthorizedKeysCommandUser' /etc/ssh/sshd_config; then
    echo "    - Adding AuthorizedKeysCommandUser to sshd_config"
    echo "AuthorizedKeysCommandUser root" >> /etc/ssh/sshd_config.
else
    echo "    - AuthorizedKeysCommandUser already set up, skipping."
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

    echo systemctl enable github-authorized-keys
    echo systemctl start github-authorized-keys
    echo systemctl restart sshd
else
    echo "Configuration is already set up, enabling and starting the service, restarting sshd."

    systemctl enable github-authorized-keys
    if systemctl --type=service --state=active | grep -Fq "github-authorized-keys.service"; then
        systemctl restart github-authorized-keys
    else
        systemctl start github-authorized-keys
    fi
    systemctl restart sshd
fi

