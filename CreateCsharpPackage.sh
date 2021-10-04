set -e

specurl=$1
version=$2
name=$3
repoOwner=$4
repoId=$5

echo Spec Url $specurl
echo Version $version
echo Name $name

mkdir -p ./csharp_service

case $specurl in
    "http"*) 
		echo Downloading spec file
		curl -k --http1.1 --connect-timeout 30 --retry 300 --retry-delay 5 --retry-connrefused $specurl > ./spec.json	;;
    *) 
		echo Copying spec file
		cp $specurl ./spec.json ;;
esac

echo Generating API
npx openapi-generator generate -i ./spec.json -g csharp-netcore -o csharp_service --git-user-id $repoOwner --git-repo-id $repoId --additional-properties=targetFramework=netcoreapp3.1,packageName=$name,packageVersion=$version,netCoreProjectFile=true

echo Copying Nuget file
cp ./nuget.config ./csharp_service/nuget.config

echo Packing Solution
cd ./csharp_service/
dotnet pack -c Release -p:Version=$version

echo Pushing Package
dotnet nuget push ./src/$name/bin/Release/**/*.nupkg -s mqube.packages --skip-duplicate