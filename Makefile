
kind-setup:
	kind create cluster --config hack/kind.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	helm repo add jetstack https://charts.jetstack.io
	helm repo update
	helm install \
		cert-manager jetstack/cert-manager \
		--namespace cert-manager \
		--create-namespace \
		--version v1.10.1 \
		--set installCRDs=true
