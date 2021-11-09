LDFLAGS:='-w -s'
REPO:=rschoof
IMAGE:=prom-hue-sensors

all: build

prep:
	mkdir -p \
		docker/out/linux/arm/v7 \
		docker/out/linux/amd64 \
		docker/out/linux/arm64

build: build_arm7 build_arm64 build_amd64

build_arm7: prep
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS) -o docker/out/linux/arm/v7/prom-hue-sensors ./cmd/prom-hue-sensors
build_arm64: prep
	GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -o docker/out/linux/arm64/prom-hue-sensors ./cmd/prom-hue-sensors
build_amd64: prep
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o docker/out/linux/amd64/prom-hue-sensors ./cmd/prom-hue-sensors

docker: build
	cd docker; \
	docker buildx build --platform linux/amd64,linux/arm/v7,linux/arm64 -t $(REPO)/$(IMAGE):latest --push .
