build-local:
	go build -o bin/kuilei github.com/airconduct/kuilei/cmd/kuilei

test:
	go test -v --race ./...

VERSION ?= $(shell git describe --always)
IMAGE_REGISTRY ?= airconduct/kuilei
GOPROXY ?= https://proxy.golang.org,direct
release:
	docker buildx build --platform=linux/amd64,linux/arm64 \
		--label=${VERSION} \
		--build-arg GOPROXY=${GOPROXY}  \
		-t ${IMAGE_REGISTRY}:${VERSION} --push .

# kind-setup:
# 	kind create cluster --config hack/kind.yaml
# 	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
# 	helm repo add jetstack https://charts.jetstack.io
# 	helm repo update
# 	helm install \
# 		cert-manager jetstack/cert-manager \
# 		--namespace cert-manager \
# 		--create-namespace \
# 		--version v1.10.1 \
# 		--set installCRDs=true
