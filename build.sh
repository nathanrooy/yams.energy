mkdir public
mkdir public/static
mkdir public/authorization-pass
mkdir public/authorization-fail

cp -rf web/static/* public/static

go run scripts/build.go