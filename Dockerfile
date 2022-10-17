FROM node:14-alpine3.16

# Alpine does not contain dpkg by default
RUN apk update && apk upgrade --ignore alpine-baselayout

# OpenAPI tools install
RUN npm install -g @openapitools/openapi-generator-cli@2.5.1
RUN npm install -g @angular/compiler-cli@13.3.11 @angular/platform-server@13.3.11 @angular/compiler@13.3.11
RUN npm install -g typescript@4.8.2

# Dotnet install
RUN apk add bash icu-libs krb5-libs libgcc libintl libssl1.1 libstdc++ zlib \
    && apk add libgdiplus --repository https://dl-3.alpinelinux.org/alpine/edge/testing/
RUN mkdir -p /usr/share/dotnet \
    && ln -s /usr/share/dotnet/dotnet /usr/bin/dotnet
RUN wget https://dot.net/v1/dotnet-install.sh
RUN chmod +x dotnet-install.sh
RUN ./dotnet-install.sh -c 6.0 --install-dir /usr/share/dotnet

# Java install
RUN apk add openjdk16
RUN wget https://services.gradle.org/distributions-snapshots/gradle-8.0-20221015221816+0000-bin.zip \
    && mkdir /opt/gradle \
    && unzip -d /opt/gradle gradle-8.0-20221015221816+0000-bin.zip \
    && rm gradle-8.0-20221015221816+0000-bin.zip
ENV PATH=$PATH:/opt/gradle/gradle-8.0-20221015221816+0000/bin

ADD openapitools.json openapitools.json
ADD configuration.ts configuration.ts
ADD CreateAngularPackageV2.sh CreateAngularPackageV2.sh
ADD CreateAngularPackageV3.sh CreateAngularPackageV3.sh
ADD CreateCsharpPackage.sh CreateCsharpPackage.sh
ADD CreateJavaPackage.sh CreateJavaPackage.sh

RUN chmod +x CreateAngularPackageV2.sh \
    && chmod +x CreateAngularPackageV3.sh \
    && chmod +x CreateCsharpPackage.sh \
    && chmod +x CreateJavaPackage.sh
