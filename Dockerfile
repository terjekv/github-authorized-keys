FROM --platform=$BUILDPLATFORM alpine:3.15.4 as stage
# FROM --platform=linux/arm64 alpine:3.15.4 as stage-arm64

ARG TARGETARCH

# FROM stage-${TARGETARCH} as final

# For production run most common user add flags
#
# We need --force-badname because github users could contains capital letters, what is not acceptable in some distributions
# Really regexp to verify badname rely on environment var that set in profile.d so we rarely hit this errors.
#
# adduser wants user name be the head and flags the tail.
ENV LINUX_USER_ADD_TPL            "useradd --password x --shell {shell} {username}"
ENV LINUX_USER_ADD_WITH_GID_TPL   "useradd --password x --shell {shell} --group {group} {username}"
ENV LINUX_USER_ADD_TO_GROUP_TPL   "adduser {username} {group}"
ENV LINUX_USER_DEL_TPL            "deluser {username}"

ENV SSH_RESTART_TPL               "/usr/bin/systemctl restart sshd.service"

ENV GITHUB_API_TOKEN=
ENV GITHUB_ORGANIZATION=
ENV GITHUB_ADMIN_TEAM_NAME=
ENV GITHUB_ADMIN_TEAM_ID=

ENV GITHUB_USER_TEAM_NAME=
ENV GITHUB_USER_TEAM_ID=

ENV ETCD_ENDPOINT=
ENV ETCD_TTL=
ENV ETCD_PREFIX=/github-authorized-keys

ENV SYNC_USERS_GID=
ENV SYNC_USERS_GROUPS=
ENV SYNC_USERS_SHELL=/bin/bash
ENV SYNC_USERS_INTERVAL=

ENV INTEGRATE_SSH=false

ENV LISTEN=":301"

# For production we run container with host network, so expose is just for testing and CI\CD
EXPOSE 301

RUN apk --update --no-cache add libc6-compat ca-certificates shadow && \
    ln -s /lib /lib64

COPY ./github-authorized-keys.${TARGETARCH} /usr/bin/github-authorized-keys
RUN chmod +x /usr/bin/github-authorized-keys

ENTRYPOINT ["github-authorized-keys"]
