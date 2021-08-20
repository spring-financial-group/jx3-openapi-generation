FROM node:latest

RUN npm install @openapitools/openapi-generator-cli@1.0.18-4.3.1 -g
RUN npm install -g @angular/compiler-cli@8.2.14 @angular/platform-server@8.2.14 @angular/compiler@8.2.14

RUN wget https://packages.microsoft.com/config/ubuntu/21.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
RUN dpkg -i packages-microsoft-prod.deb
RUN rm packages-microsoft-prod.deb
RUN apt-get update; apt-get install -y apt-transport-https && apt-get update && apt-get install -y dotnet-sdk-3.1

RUN apt-get update
RUN apt-get -y install default-jre

ADD configuration.ts configuration.ts
ADD CreateAngularPackageV2.sh CreateAngularPackageV2.sh
ADD CreateAngularPackageV3.sh CreateAngularPackageV3.sh
ADD CreateCsharpPackage.sh CreateCsharpPackage.sh

RUN chmod +x CreateAngularPackageV2.sh
RUN chmod +x CreateAngularPackageV3.sh
RUN chmod +x CreateCsharpPackage.sh
