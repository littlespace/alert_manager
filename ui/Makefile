DOCKER_IMAGE=chi1-docker-registry.simulprod.com/alert-manager-ui
DOCKER_TAG=v0.1.3

build: build-prod build-integration

build-prod: 
	docker build \
		--build-arg alert_manager_server="https://alert-manager-api.simulprod.com/" \
		 -t $(DOCKER_IMAGE):$(DOCKER_TAG)-prod .

build-integration: 
	docker build \
		--build-arg alert_manager_server="https://alert-manager-integration-api.simulprod.com/" \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG)-integration .
