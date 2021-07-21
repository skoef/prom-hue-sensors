all: build

prep:
	mkdir -p out/linux/arm/v7 out/linux/amd64

build: build_arm build_amd64

build_arm: prep
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags '-w -s' -o out/linux/arm/v7/tempread ./cmd/tempread

build_amd64: prep
	GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o out/linux/amd64/tempread ./cmd/tempread

docker:
	docker buildx build --platform linux/amd64,linux/arm/v7 -t rschoof/tempread --push .
