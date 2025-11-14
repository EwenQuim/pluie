css:
	tailwindcss -i ./src/input.css -o ./static/tailwind.min.css --watch --minify

build: css
	go build -v -ldflags="-s -w" -o pluie-app

# Build for local testing
docker-build:
	docker build --platform linux/amd64 -t ghcr.io/ewenquim/pluie:latest .

docker-preview:
	docker run --env-file .env --rm -p 9999:9999 ghcr.io/ewenquim/pluie:latest
