
docker build --pull --rm -f "Dockerfile_royal" -t dockerklint/aim_royal:v1.4 "."

docker build --pull --rm -f "Dockerfile_gateway" -t dockerklint/aim_gateway:v1.4 "."

docker build --pull --rm -f "Dockerfile_server" -t dockerklint/aim_server:v1.4 "."

docker build --pull --rm -f "Dockerfile_router" -t dockerklint/aim_router:v1.1 "."

docker push dockerklint/aim_royal:v1.4
docker push dockerklint/aim_gateway:v1.4
docker push dockerklint/aim_server:v1.4
docker push dockerklint/aim_router:v1.1