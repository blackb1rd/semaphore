FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS builder

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    nodejs \
    npm \
    git \
    gcc \
    zip \
    unzip \
    tar \
    ca-certificates; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /usr/local
# hadolint ignore=DL4006
RUN curl -sL https://taskfile.dev/install.sh | sh

WORKDIR /go/src/semaphore
COPY . /go/src/semaphore

RUN --mount=type=cache,target=/go/pkg \
    go mod download -x

ARG APP_BUILD_TYPE
ARG TARGETOS
ARG TARGETARCH
ARG GH_TOKEN

RUN if [ -n "$APP_BUILD_TYPE" ]; then \
    git clone https://${GH_TOKEN}@github.com/semaphoreui/semaphorepro-module.git pro_impl && \
    go work init . ./pro_impl; \
    fi

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task deps APP_BUILD_TYPE=${APP_BUILD_TYPE} && \
    task build GOOS=${TARGETOS} GOARCH=${TARGETARCH} APP_BUILD_TYPE=${APP_BUILD_TYPE}


ENV OPENTOFU_VERSION="1.9.0"
ENV TERRAFORM_VERSION="1.11.3"
ENV TERRAGRUNT_VERSION="0.78.0"
#ENV PULUMI_VERSION="3.116.1"
#ENV POWERSHELL_VERSION="3.116.1"

RUN wget https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz && \
    tar xf tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz -C /tmp && \
    rm tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz

RUN curl -O https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip -d /tmp && \
    rm terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip

RUN wget -O /tmp/terragrunt https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_${TARGETARCH} && \
    chmod +x /tmp/terragrunt

FROM debian:bookworm

ARG TARGETARCH="amd64"
# renovate: datasource=pypi depName=ansible
ARG ANSIBLE_VERSION=11.1.0
ENV ANSIBLE_VERSION=${ANSIBLE_VERSION}
ARG ANSIBLE_VENV_PATH=/opt/semaphore/apps/ansible/${ANSIBLE_VERSION}/venv

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    bash \
    curl \
    git \
    gnupg \
    default-mysql-client \
    openssh-client \
    python3 \
    python3-venv \
    python3-pip \
    rsync \
    sshpass \
    tar \
    tini \
    tzdata \
    unzip \
    wget \
    zip \
    jq \
    ca-certificates \
    nodejs \
    npm \
    chromium \
    nmap \
    libgtk-3-0 \
    libnotify4 \
    libgconf-2-4 \
    libnss3 \
    libxss1 \
    libasound2 \
    xvfb \
    fonts-liberation \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libx11-xcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxi6 \
    libxrandr2 \
    libxtst6 \
    libpangocairo-1.0-0 \
    libpango-1.0-0 \
    libdbus-1-3; \
    rm -rf /var/lib/apt/lists/*; \
    groupadd -g 0 root || true; \
    useradd -u 1001 -g 0 -M -s /bin/sh semaphore || true; \
    mkdir -p /tmp/semaphore /etc/semaphore /var/lib/semaphore /opt/semaphore; \
    chown -R semaphore:0 /tmp/semaphore /etc/semaphore /var/lib/semaphore /opt/semaphore

RUN echo $'Host *\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null' > /etc/ssh/ssh_config.d/semaphore.conf

COPY ./deployment/docker/server/ansible.cfg /etc/ansible/ansible.cfg
COPY --from=builder /go/src/semaphore/deployment/docker/server/server-wrapper /usr/local/bin/
COPY --from=builder /go/src/semaphore/bin/semaphore /usr/local/bin/
COPY --from=builder /tmp/tofu /usr/local/bin/
COPY --from=builder /tmp/terraform /usr/local/bin/
COPY --from=builder /tmp/terragrunt /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/server-wrapper && \
    chmod +x /usr/local/bin/server-wrapper && \
    chown -R semaphore:0 /usr/local/bin/semaphore && \
    chmod +x /usr/local/bin/semaphore

WORKDIR /home/semaphore

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends python3-dev build-essential libssl-dev libffi-dev cargo; \
    mkdir -p ${ANSIBLE_VENV_PATH}; \
    python3 -m venv ${ANSIBLE_VENV_PATH} --system-site-packages; \
    . ${ANSIBLE_VENV_PATH}/bin/activate; \
    pip3 install --upgrade pip ansible==${ANSIBLE_VERSION} boto3 botocore requests pywinrm passlib docker; \
    # install community collections so their modules/plugins are available
    ansible-galaxy collection install community.docker; \
    ansible-galaxy collection install community.general; \
    ansible-galaxy collection install amazon.aws; \
    apt-get remove -y --purge python3-dev build-essential libssl-dev libffi-dev cargo; \
    rm -rf /var/lib/apt/lists/*; \
    find ${ANSIBLE_VENV_PATH} -iname __pycache__ | xargs -r rm -rf; \
    chown -R semaphore:0 /opt/semaphore

# Do not install Docker engine inside the image. For Ansible docker modules
# prefer mounting the host Docker socket at runtime: -v /var/run/docker.sock:/var/run/docker.sock

# Install Cypress globally (downloads binary during install)
RUN set -eux; \
    npm install -g --unsafe-perm cypress@15 || npm install -g --unsafe-perm cypress || true; \
    # run install to ensure Cypress binary is fetched for the target platform
    npx cypress install || true

USER 1001
EXPOSE 3000

ENV VIRTUAL_ENV="$ANSIBLE_VENV_PATH"
ENV PATH="$ANSIBLE_VENV_PATH/bin:$PATH"

# Preventing ansible zombie processes. Tini kills zombies.
ENTRYPOINT ["/usr/bin/tini", "--"]
CMD [ "/usr/local/bin/server-wrapper"]
