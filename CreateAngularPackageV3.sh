set -e

specurl=$1
packageversion=$2

echo Spec Url $specurl
echo Version $packageversion

mkdir -p ./service

case $specurl in
    "http"*) 
		echo Downloading spec file
		curl --connect-timeout 30 --retry 300 --retry-delay 5 --retry-connrefused $specurl > ./spec.json	;;
    *) 
		echo Copying spec file
		cp $specurl ./spec.json ;;
esac

npx openapi-generator-cli generate -i ./spec.json -g typescript-angular -o service --additional-properties=fileNaming=camelCase --enable-post-process-file

sed -i 's/0.0.0/'$packageversion'/g' ./package.json

cp ./package.json ./service/package.json 
cp ./tsconfig.json ./service/tsconfig.json 
cp ./configuration.ts ./service/configuration.ts

cd service

npm install --save rxjs@6.6.7
npm install --save zone.js@0.9.1
npm install --save @angular/core@8.2.14 
npm install --save @angular/common@8.2.14 
npm install --save typescript

ngc 

cd ..

cp ./.npmrc ./dist/.npmrc
cp ./package.json ./dist/package.json

cd dist

npm publish
