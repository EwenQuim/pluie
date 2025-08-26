css:
	tailwindcss -i ./src/input.css -o ./static/tailwind.min.css --watch --minify

build: css
	go build -v -ldflags="-s -w" -o pluie-app

# Build for Prod
docker-build:
	docker build --platform linux/amd64 -t ewenquim/pluie .

docker-push:
	docker push ewenquim/pluie

docker-build-and-push: docker-build docker-push

docker-preview:
	docker run --env-file .env --rm -p 9999:9999 ewenquim/pluie
