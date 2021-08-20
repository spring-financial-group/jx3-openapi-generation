set -e

specurl=$1
version=$2
name=$3

echo Spec Url $specurl
echo Version $version
echo Name $version

mkdir -p ./csharp_service

case $specurl in
    "http"*) 
		echo Downloading spec file
		curl --connect-timeout 30 --retry 300 --retry-delay 5 --retry-connrefused $specurl > ./spec.json	;;
    *) 
		echo Copying spec file
		cp $specurl ./spec.json ;;
esac

openapi-generator generate -i ./spec.json -g csharp-netcore -o csharp_service --additional-properties=targetFramework=netcoreapp3.1,packageName=$name,packageVersion=$version

cp ./nuget.config ./csharp_service/nuget.config

cd ./csharp_service/
dotnet pack -c Release

dotnet nuget push ./src/$name/bin/Release/**/*.nupkg -s mqube.packages --skip-duplicate