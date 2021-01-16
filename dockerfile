FROM mcr.microsoft.com/dotnet/core/sdk:3.1-alpine

RUN dotnet tool install --global --version 5.6.2 Swashbuckle.AspNetCore.Cli