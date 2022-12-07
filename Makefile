VERSION ?= $(shell cat version)
IMAGE_REGISTRY ?= airconduct/kuilei

publish:
	npm run build
	docker buildx build --platform=linux/amd64,linux/arm64  -t ${IMAGE_REGISTRY}:${VERSION} --push .
