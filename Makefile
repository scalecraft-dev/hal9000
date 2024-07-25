.PHONY: run
run:
	docker compose -f .docker/docker-compose.yml up -d

.PHONY: stop
stop:
	docker compose -f .docker/docker-compose.yml down

.PHONY: build-docs
build-docs:
	swagger generate spec -m -i docs/swagger-tags.json -o docs/swagger.yaml
	redocly build-docs docs/swagger.yaml -o docs/index.html

.PHONY: build-image
build-image:
	docker build \
		-f .docker/carestat.Dockerfile \
		-t $(AWS_ACCOUNT_ID).dkr.ecr.us-east-1.amazonaws.com/data/carestat:latest .

.PHONY: push-image
push-image:
	docker push $(AWS_ACCOUNT_ID).dkr.ecr.us-east-1.amazonaws.com/data/carestat:latest

.PHONY: login-docker
login-docker:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.us-east-1.amazonaws.com
