FROM node:alpine

RUN npm install @openapitools/openapi-generator-cli@1.0.18-4.3.1 -g
RUN npm install -g @angular/compiler-cli@8.2.14 @angular/platform-server@8.2.14 @angular/compiler@8.2.14