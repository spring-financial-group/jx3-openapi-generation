set -e

specurl=$1
version=$2
name=$3
repoOwner=$4
repoId=$5

echo Spec Url $specurl
echo Version $version
echo Name $name

mkdir -p ./java_service

case $specurl in
    "http"*)
		echo Downloading spec file
		curl -k --http1.1 --connect-timeout 30 --retry 300 --retry-delay 5 --retry-connrefused $specurl > ./spec.json	;;
    *)
		echo Copying spec file
		cp $specurl ./spec.json ;;
esac

cp /openapitools.json ./openapitools.json

echo Generating API
npx openapi-generator generate -i ./spec.json -g java -o java_service --git-user-id $repoOwner --git-repo-id $repoId --global-property models,modelTests=false,modelDocs=false -p basePackage=mqube.caseService -p modelPackage=mqube.caseService.models -p dateLibrary=java8

echo Copying Gradle file
cp ./build.gradle ./java_service/build.gradle

echo Pushing Package
gradle publish