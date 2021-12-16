FROM node:14

RUN npm install -g @openapitools/openapi-generator-cli@2.3.9
RUN npm install -g @angular/compiler-cli@8.2.14 @angular/platform-server@8.2.14 @angular/compiler@8.2.14
RUN npm install -g typescript@3.4.5

RUN wget https://packages.microsoft.com/config/ubuntu/21.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb \
    && dpkg -i packages-microsoft-prod.deb \
    && rm packages-microsoft-prod.deb

RUN apt-get update && apt-get install -y apt-transport-https dotnet-sdk-3.1 default-jre zip

RUN wget https://services.gradle.org/distributions/gradle-7.3.2-bin.zip \
    && mkdir /opt/gradle \
    && unzip -d /opt/gradle gradle-7.3.2-bin.zip \
    && rm gradle-7.3.2-bin.zip

ENV PATH=$PATH:/opt/gradle/gradle-7.3.2/bin

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
