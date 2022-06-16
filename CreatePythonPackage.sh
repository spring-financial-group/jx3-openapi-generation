set -e

specPath=$1
serviceName=$2
version=$3
gitToken=$3

gitUrl=https://mqube-bot:"$gitToken"@github.com/spring-financial-group/mqube-ml-doc-pipeline-schemas.git

git clone "$gitUrl"
cd mqube-ml-doc-pipeline-schemas
git remote set-url origin "$gitUrl"

echo Generating python package
datamodel-codegen  --input "$specPath" --input-file-type openapi --output "$serviceName".py

git add "$serviceName".py
git commit -m "chore(deps): upgrade $serviceName.py -> v$version"
git push origin master