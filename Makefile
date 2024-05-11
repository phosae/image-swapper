DOCKER_USER=zengxu
TAG = $(DOCKER_USER)/image-swapper:v1

install:
	make ensure-image
	helm install image-swapper --namespace kube-system ./charts

uninstall:
	helm uninstall image-swapper --namespace kube-system 

ensure-image: 
ifeq ([], $(shell docker inspect --type=image $(TAG)))
	make docker-build-load
endif

docker-build-load:
	docker buildx build --load -t $(TAG) .
	kind load docker-image $(TAG)
