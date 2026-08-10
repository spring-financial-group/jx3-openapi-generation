FROM jx3mqubebuild.azurecr.io/spring-financial-group/all-sdks-debian:latest

ENV GOPATH="/usr/local/go"
RUN go install github.com/vektra/mockery/v3@v3.7.2

# Install javascript dependencies
RUN npm install -g \
    @openapitools/openapi-generator-cli@2.23.1 \
    @angular/compiler-cli@13.3.1 \
    @angular/platform-server@13.3.1 \
    @angular/compiler@13.3.1 \
    typescript@4.6.3

## Copy CLI binary & add to PATH
COPY ./build/linux /jx3-openapi-generation
ENV PATH "$PATH:/jx3-openapi-generation"