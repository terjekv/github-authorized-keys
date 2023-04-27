#!/bin/bash

set -e

declare -A env_vars
declare -a env_order

shopt -s nocasematch

cd /tmp

GAK_VERSION=0.0.25
UNAME_ARCH=$(uname -m)
BRANCH="${1:-main}"

CURL="curl -sSfOL"
BINARY_PATH="/usr/local/sbin"
# Loading a given configuration file from --config doesn't work, so we use the hardcoded default
# path for the configuration file...
CONF_FILE="$HOME/.github-authorized-keys.yaml" 
SELINUX_POLICIES="github-authorized-keys-allow-sshd-reserved-ports.pp"

# https://raw.githubusercontent.com/terjekv/github-authorized-keys/main/contrib/
RAW_CONTRIB_URL=https://raw.githubusercontent.com/terjekv/github-authorized-keys/${BRANCH}/contrib

# Default environment file, is overridden by install_system_specific_artifacts
ENV_FILE=env

header() {
    echo
    echo "** ${1} **"
    echo 
}

display() {
    # space before the string is $2 * 2 spaces, minimum 2
    printf "%$(( ${2:-1} * 2 ))s - %s\n" " " "${1}"
}

load_env_vars() {
    ENV_FILE=$1
    while IFS="=" read -r key value; do
        if [[ ! "${key}" =~ ^# ]] && [[ -n "${key}" ]]; then
            value="${value%\"}"
            value="${value#\"}"
            env_vars["${key}"]="${value}"
            default_env_vars["${key}"]="${value}"
            env_order+=("${key}")
        fi
    done < "${ENV_FILE}"

    if [[ -f ${CONF_FILE} ]]; then
        while IFS=": " read -r key value; do
            if [[ ! "${key}" =~ ^# ]] && [[ -n "${key}" ]]; then
                key="${key// /}"
                value="${value%\"}"
                value="${value#\"}"
                env_vars["${key}"]="${value}"
            fi
        done < ${CONF_FILE}
    fi
}

get_default_value() {
    local token=$1
    local vname="GAK_${token}"
    local val=${!vname}

    if [[ -z "${val}" ]]; then
        val=${env_vars["${token}"]}
    fi

    echo "${val}"
}

print_and_confirm() {
    local output=""
    for token in "${env_order[@]}"; do
        val=$(get_default_value "${token}")

        if [[ -n "${GAK_INTERACTIVE}" ]]; then
            read -p "${token} [${val}]: " interactive_val
            if [[ -n "${interactive_val}" ]]; then
                val=${interactive_val}
            fi
            # Update the value in env_vars for the next run
            env_vars["${token}"]="${val}"
        fi

        output+="${token}: \"${val}\""$'\n'
    done

    if [[ -n "${GAK_INTERACTIVE}" ]]; then
        echo "---------- configuration ------------"
        printf "%s" "${output}"
        echo "-------------------------------------"
        read -p "Does this look okay? (yes/ok/enter to confirm) " confirm
        if [[ ! "${confirm}" =~ ^(yes|ok)?$ ]]; then
            print_and_confirm
            return
        fi
    fi

    local file_operation="created"
    if [[ -f ${CONF_FILE} ]]; then
        file_operation="updated"
    fi

    echo "---" > ${CONF_FILE}
    printf "%s" "${output}" >> ${CONF_FILE}
    display "${CONF_FILE} has been ${file_operation}..." 2
}

fix_selinux_contexts() {
    if [[ "${ID_LIKE}" =~ "fedora" ]]; then
        display "Ensuring SELinux contexts are correct."

        # A note on the code below.
        # We use grep > /dev/null rather than grep -q to prevent
        # BrokenPipeError: [Errno 32] Broken pipe
        # as a result of the pipe being closed by grep on the first hit.
        for binary in authorized-keys github-authorized-keys; do
            display "${BINARY_PATH}/${binary}" 2
            # If a policy exists for the binary, ensure we have the latest version by
            # first removing it...
            /usr/sbin/semanage fcontext -l | grep ${BINARY_PATH}/${binary} > /dev/null && \
                /usr/sbin/semanage fcontext -d ${BINARY_PATH}/${binary}
            # ...then adding the correct one...
            /usr/sbin/semanage fcontext -a -t bin_t ${BINARY_PATH}/${binary}
            # ...and finally relabeling the binary.
            /usr/sbin/restorecon ${BINARY_PATH}/${binary}
        done
    #    chcon system_u:object_r:bin_t:s0 ${BINARY_PATH}/authorized-keys ${BINARY_PATH}/github-authorized-keys
    fi
}

validate_environment() {
    if [ "$(id -u)" -ne 0 ]; then
            echo 'This script must be run by root' >&2
            exit 1
    fi

    if [ "${UNAME_ARCH}" == "x86_64" ]; then
        ARCH="amd64"
    elif [ "${UNAME_ARCH}" == "aarch64" ]; then
        ARCH="arm64"
    else
        echo "Unsupported architecture: ${UNAME_ARCH}"
        exit 1
    fi
}

fetch_artifacts() {
    # Fetch shared artifacts
    $CURL "https://github.com/terjekv/github-authorized-keys/releases/download/v${GAK_VERSION}/github-authorized-keys-v${GAK_VERSION}-linux-${ARCH}.tar.gz"
    $CURL "${RAW_CONTRIB_URL}/authorized-keys"

    # Fetch systemd service file
    $CURL "${RAW_CONTRIB_URL}/github-authorized-keys.service"
}

preamble() {
    [ -f /etc/os-release ] && source /etc/os-release

    header "Installing github-authorized-keys v${GAK_VERSION}..."
    display "Detected OS: ${PRETTY_NAME} on ${ARCH}"
}

install_system_specific_artifacts() {
    if [[ "${ID_LIKE}" =~ "fedora" ]]; then
        display "System is Fedora-like, expecting systemd and SELinux."
        display "Installing SELinux policies."

        # SElinux
        for policy in $SELINUX_POLICIES; do
            display "Installing ${policy}" 2
            $CURL "${RAW_CONTRIB_URL}/${policy}"
            semodule -i ${policy}
            rm ${policy}
        done

        $CURL "${RAW_CONTRIB_URL}/env.rhel"
        ENV_FILE=env.rhel
    else
        display "No specific setup available, assuming debian-like with systemd but no SELinux."

        $CURL "${RAW_CONTRIB_URL}/env"
        ENV_FILE=env
    fi
}

install_generic_artifacts() {
    # Unpack binary
    tar xzf github-authorized-keys-v${GAK_VERSION}-linux-${ARCH}.tar.gz
    rm github-authorized-keys-v${GAK_VERSION}-linux-${ARCH}.tar.gz README.md LICENSE

    display "Installing binary files into ${BINARY_PATH}."
    # Move artifacts into place
    for file in authorized-keys github-authorized-keys; do
        display "Installing ${file}." 2
        mv ${file} ${BINARY_PATH}/
    done

    display "Installing systemd service."
    mv github-authorized-keys.service /etc/systemd/system/github-authorized-keys.service
}

fix_binary_permissions() {
    display "Fixing permissions for binaries."
    for binary in authorized-keys github-authorized-keys; do
        display "${BINARY_PATH}/${binary}." 2
        chown root:root ${BINARY_PATH}/${binary}
        chmod 755 ${BINARY_PATH}/${binary}
    done
}

create_configuration_file() {
    display "Creating configuration file ${CONF_FILE}."
    load_env_vars $ENV_FILE
    print_and_confirm
}

ensure_ssh_keys_group() {
    display "Ensuring default group exists."
    if ! grep -Fq ':999:' /etc/group; then
        groupadd -g 999 ssh_keys
    fi
}

validate_ssh_configuration() {
    ensure_ssh_keys_group

    display "Validating ssh configuration."

    if ! grep -Eq '^AuthorizedKeysCommand' /etc/ssh/sshd_config; then
        display "Adding AuthorizedKeysCommand to sshd_config." 2
        echo "AuthorizedKeysCommand ${BINARY_PATH}/authorized-keys" >> /etc/ssh/sshd_config 
    else
        display "AuthorizedKeysCommand already set up, skipping." 2
    fi

    if ! grep -Eq '^AuthorizedKeysCommandUser' /etc/ssh/sshd_config; then
        display "Adding AuthorizedKeysCommandUser to sshd_config." 2
        echo "AuthorizedKeysCommandUser root" >> /etc/ssh/sshd_config
    else
        display "AuthorizedKeysCommandUser already set up, skipping." 2
    fi

    # https://github.com/widdix/aws-ec2-ssh/issues/157
    display "Ensuring that ec2-instance-connect is not installed."
    if [ -f /usr/bin/apt-get ]; then
        /usr/bin/apt-get -qq remove ec2-instance-connect > /dev/null
    fi
}

finish() {
    config_token=$(grep -E '^GITHUB_API_TOKEN' $CONF_FILE | awk -F: '{print $2}' | cut -f2 -d'"')
    if [ "${config_token}" == "${default_env_vars[GITHUB_API_TOKEN]}" ]; then
        display "Default github token found in config file, not starting services."
    else
        sctl="/usr/bin/systemctl --quiet"
        display "Reloading systemd daemon and enabling/restarting services."
        $sctl daemon-reload

        $sctl enable github-authorized-keys
        if /usr/bin/systemctl --type=service --state=active | grep -Fq "github-authorized-keys.service"; then
            $sctl restart github-authorized-keys
        else
            $sctl start github-authorized-keys
        fi
        $sctl restart sshd
    fi

    header "Installation complete!"
}

#
# Start of script
#

validate_environment

preamble

fetch_artifacts

install_system_specific_artifacts
install_generic_artifacts

fix_binary_permissions
fix_selinux_contexts

validate_ssh_configuration

create_configuration_file

finish
