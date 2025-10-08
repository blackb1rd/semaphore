ARG SEMAPHORE_IMAGE
ARG SEMAPHORE_VERSION

FROM ${SEMAPHORE_IMAGE}:${SEMAPHORE_VERSION}

ARG TARGETARCH
ARG POWERSHELL_VERSION="7.5.0"

USER root

# Install dependencies and download PowerShell for Debian Bookworm
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    less \
    ncurses-term \
    krb5-user \
    libgcc1 \
    gettext-base \
    libssl3 \
    libstdc++6 \
    tzdata \
    userspace-rcu \
    zlib1g \
    libicu-dev \
    curl \
    lttng-tools \
    openssh-client; \
    rm -rf /var/lib/apt/lists/*; \
    if [ "${TARGETARCH}" = "amd64" ]; then PS_NAME="linux-x64"; elif [ "${TARGETARCH}" = "arm64" ]; then PS_NAME="linux-arm64"; else PS_NAME="linux-x64"; fi; \
    wget -O /tmp/powershell.tar.gz "https://github.com/PowerShell/PowerShell/releases/download/v${POWERSHELL_VERSION}/powershell-${POWERSHELL_VERSION}-${PS_NAME}.tar.gz"; \
    mkdir -p /opt/microsoft/powershell/${POWERSHELL_VERSION}; \
    tar zxf /tmp/powershell.tar.gz -C /opt/microsoft/powershell/${POWERSHELL_VERSION}; \
    rm /tmp/powershell.tar.gz; \
    chmod +x /opt/microsoft/powershell/${POWERSHELL_VERSION}/pwsh; \
    ln -s /opt/microsoft/powershell/${POWERSHELL_VERSION}/pwsh /usr/local/bin/pwsh; \
    ln -s /opt/microsoft/powershell/${POWERSHELL_VERSION}/pwsh /usr/local/bin/powershell

USER 1001