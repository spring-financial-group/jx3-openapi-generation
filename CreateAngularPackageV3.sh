set -e

specurl=$1

echo Spec Url $specurl

cd registry

mkdir -p ./service

case $specurl in
    "http"*) 
		echo Downloading spec file
		curl --http1.1 --connect-timeout 30 --retry 300 --retry-delay 5 --retry-connrefused $specurl > ./spec.json
		;;
    *) 
		echo Copying spec file
		cp $specurl ./spec.json ;;
esac


cp /openapitools.json ./openapitools.json 

npx openapi-generator-cli generate -i ./spec.json -g typescript-angular -o service --additional-properties=fileNaming=camelCase,ngVersion=10.0.0 --enable-post-process-file --remove-operation-id-prefix

cp ./package.json ./service/package.json 
cp ./tsconfig.json ./service/tsconfig.json 
cp /configuration.ts ./service/configuration.ts

cd service

npm install --save rxjs@6.6.7
npm install --save zone.js@0.9.1
npm install --save @angular/core@8.2.14 
npm install --save @angular/common@8.2.14

ngc 

cd ..

cp ./.npmrc ./dist/.npmrc
cp ./package.json ./dist/package.json

cd dist

npm publish
