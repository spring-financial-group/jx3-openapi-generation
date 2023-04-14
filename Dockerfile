FROM node:14

RUN apt-get update

# Install Go (1.20.3)
RUN wget -c https://dl.google.com/go/go1.20.3.linux-amd64.tar.gz -O - | tar -xz -C /usr/local
ENV PATH "$PATH:/usr/local/go/bin"
RUN go version

# Install python dependencies
RUN apt-get install python3-pip -y && pip3 install datamodel-code-generator

# Install node dependencies
RUN npm install -g @openapitools/openapi-generator-cli@2.4.6
RUN npm install -g @angular/compiler-cli@13.3.1 @angular/platform-server@13.3.1 @angular/compiler@13.3.1
RUN npm install -g typescript@4.6.3

# Install dotnet dependencies
RUN wget https://packages.microsoft.com/config/ubuntu/21.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb \
    && dpkg -i packages-microsoft-prod.deb \
    && rm packages-microsoft-prod.deb

RUN apt-get update && apt-get install -y apt-transport-https dotnet-sdk-3.1 openjdk-11-jdk default-jre zip

# Install java dependencies
RUN wget https://services.gradle.org/distributions/gradle-7.3.2-bin.zip \
    && mkdir /opt/gradle \
    && unzip -d /opt/gradle gradle-7.3.2-bin.zip \
    && rm gradle-7.3.2-bin.zip

ENV PATH=$PATH:/opt/gradle/gradle-7.3.2/bin

## Copy CLI binary & add to PATH
COPY ./build/linux /jx3-openapi-generation
ENV PATH "$PATH:/jx3-openapi-generation"

# Add pipeline scripts & config files
ADD openapitools.json openapitools.json
ADD configuration.ts configuration.ts
ADD CreateAngularPackageV2.sh CreateAngularPackageV2.sh
ADD CreateAngularPackageV3.sh CreateAngularPackageV3.sh
ADD CreateCsharpPackage.sh CreateCsharpPackage.sh
ADD CreateJavaPackage.sh CreateJavaPackage.sh
ADD FindOpenAPISpec.sh FindOpenAPISpec.sh

# Set permissions
RUN chmod +x CreateAngularPackageV2.sh \
    && chmod +x CreateAngularPackageV3.sh \
    && chmod +x CreateCsharpPackage.sh \
    && chmod +x CreateJavaPackage.sh \
    && chmod +x FindOpenAPISpec.sh
