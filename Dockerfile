FROM ghcr.io/astral-sh/uv:python3.10-bookworm

RUN apt-get update

ENV NODE_VERSION=24.13.0

# Install NodeJS
RUN wget -qO- https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-x64.tar.xz | tar -xJ -C /usr/local --strip-components=1

RUN node --version
RUN npm --version

# Install Go
RUN wget -c https://dl.google.com/go/go1.24.4.linux-amd64.tar.gz -O - | tar -xz -C /usr/local
ENV GOPATH "/usr/local/go"
ENV PATH "$PATH:$GOPATH/bin"
RUN go install github.com/vektra/mockery/v3@v3.6.1
RUN go version

# Install javascript dependencies
RUN npm install -g \
    @openapitools/openapi-generator-cli@2.23.1 \
    @angular/compiler-cli@13.3.1 \
    @angular/platform-server@13.3.1 \
    @angular/compiler@13.3.1 \
    typescript@4.6.3

# Install dotnet dependencies
RUN wget https://packages.microsoft.com/config/ubuntu/21.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb \
    && dpkg -i packages-microsoft-prod.deb \
    && rm packages-microsoft-prod.deb

RUN apt-get update && apt-get install -y apt-transport-https dotnet-sdk-7.0 openjdk-17-jdk default-jre zip

# Install java dependencies
RUN wget https://services.gradle.org/distributions/gradle-7.3.2-bin.zip \
    && mkdir /opt/gradle \
    && unzip -d /opt/gradle gradle-7.3.2-bin.zip \
    && rm gradle-7.3.2-bin.zip

ENV PATH=$PATH:/opt/gradle/gradle-7.3.2/bin

## Copy CLI binary & add to PATH
COPY ./build/linux /jx3-openapi-generation
ENV PATH "$PATH:/jx3-openapi-generation"

# Copy packaging templates
COPY ./templates /templates

# Copy individual language configuration files
COPY ./configs /configs
