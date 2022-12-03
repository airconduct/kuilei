VERSION ?= latest
IMAGE_REGISTRY ?= airconduct/kuilei

publish:
	docker buildx build --platform=linux/amd64,linux/arm64  -t ${IMAGE_REGISTRY}:${VERSION} --push .
