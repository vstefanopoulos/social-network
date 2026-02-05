NAMESPACE=social-network

.PHONY: build-base apply-monitoring apply-cors apply-kafka delete-volumes build-cnpg build-all apply-namespace apply-pvc apply-db1 apply-db2 build-services deploy-users run-migrations logs-users logs-db deploy-all reset

# === Utils ===

build-proto:
	$(MAKE) -f backend/shared/proto/protoMakefile generate

# ~~~~~~~~~~~~~~~~~~~~~~~
# ==== Docker ====


# --- Image Creation ---
build-base:
	docker build -t social-network/go-base -f backend/docker/go/base2.Dockerfile .

build-services:
	docker build -f backend/docker/services/api-gateway.Dockerfile -t social-network/api-gateway:dev .
	docker build -f backend/docker/services/chat.Dockerfile -t social-network/chat:dev .
	docker build -f backend/docker/services/live.Dockerfile -t social-network/live:dev .
	docker build -f backend/docker/services/media.Dockerfile -t social-network/media:dev .
	docker build -f backend/docker/services/notifications.Dockerfile -t social-network/notifications:dev .
	docker build -f backend/docker/services/posts.Dockerfile -t social-network/posts:dev .
	docker build -f backend/docker/services/users.Dockerfile -t social-network/users:dev .
	docker build -f backend/docker/front/front.Dockerfile -t social-network/front:dev .

# apparently minikube neesd to load the images
load-images:
	minikube image load social-network/api-gateway:dev
	minikube image load social-network/chat:dev
	minikube image load social-network/live:dev
	minikube image load social-network/media:dev
	minikube image load social-network/notifications:dev
	minikube image load social-network/posts:dev
	minikube image load social-network/users:dev
	minikube image load social-network/front:dev

build-cnpg:
	docker buildx bake -f backend/docker/cnpg/bake.hcl postgres16-cloud-native

# --- deploy from docker ---

docker-up:
	$(MAKE) create-network
	$(MAKE) build-base
	docker-compose up --build

delete-volumes:
	docker compose down
	docker volume rm backend_users-db-data backend_posts-db-data backend_chat-db-data backend_notifications-db-data backend_media-db-data

docker-test:
	$(MAKE) create-network
	docker compose -f docker-test.yml up --build

docker-up-test:
	$(MAKE) docker-up
	$(MAKE) docker-test

api:
	docker-compose up api-gateway --build

# this network is used to let the tester connect to the rest of the containers
# since the tester won't be part of the normal docker-compose, it will have it's own so that someone gotta go out of their way to test
create-network:
	@docker network inspect social-network >nul 2>&1 || docker network create social-network


# ~~~~~~~~~~~~~~~~~~
# ==== K8s ====


# --- Preliminary ---
# install cnpg operator
op-manifest:
	kubectl apply --server-side -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.28/releases/cnpg-1.28.0.yaml


full-start:
	$(MAKE) build-all; sleep 1
	$(MAKE) apply-kafka; sleep 1
	$(MAKE) apply-namespace; sleep 1
	$(MAKE) apply-configs; sleep 1
	$(MAKE) apply-pvc; sleep 1
	$(MAKE) apply-monitoring; sleep 10
	$(MAKE) apply-db1; sleep 20
	$(MAKE) apply-db2; sleep 20
	$(MAKE) run-migrations; sleep 10
	$(MAKE) apply-apps; sleep 20
	$(MAKE) apply-cors; sleep 5
	$(MAKE) port-forward; sleep 1

# --- Deployment Order ---

#-1. CHANGE TO BASH TERMINAL!

# 0. Builds all nessary docker images, Do only once or if chances happen
build-all:
	$(MAKE) build-base 
	$(MAKE) op-manifest 
	$(MAKE) build-cnpg 
	$(MAKE) build-services 
	$(MAKE) load-images 

# 1.
apply-kafka:
	kubectl create namespace kafka
	kubectl apply -f "https://strimzi.io/install/latest?namespace=kafka" -n kafka

# 1.1
apply-namespace:
	kubectl apply -f backend/k8s/ --recursive --selector stage=namespace

# 2.
apply-configs:
	kubectl apply -R -f backend/k8s/ --recursive --selector stage=config

# 2.1
apply-pvc:
	kubectl apply -f backend/k8s/ --recursive --selector stage=pvc

# 3.
apply-monitoring:
	kubectl apply -R -f backend/k8s/ --recursive --selector stage=monitoring

# 4. (only in production)
deploy-nginx:
	helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
	helm repo update
	helm upgrade --install nginx-ingress ingress-nginx/ingress-nginx \
		-n ingress-nginx --create-namespace

# 5. CNPG dbs + Redis
apply-db1:
	kubectl apply -f backend/k8s/ --recursive --selector stage=db1

#5.5 redis sentinels
apply-db2:
	kubectl apply -f backend/k8s/ --recursive --selector stage=db2

# !!! WAIT HERE !!!

# 6.
run-migrations:
	kubectl apply -f backend/k8s/ --recursive --selector stage=migration


# 7.
apply-apps:
	kubectl apply -f backend/k8s/ --recursive --selector stage=app

# 8.
apply-cors:
	kubectl apply -f backend/k8s/ --recursive --selector stage=cors

# 9.
port-forward:
	@echo "Starting port-forwards..."
	@bash -c '\
		trap "echo Cleaning up port-forwards; pkill -f \"kubectl port-forward\"" EXIT SIGINT; \
		kubectl port-forward -n frontend svc/nextjs-frontend 3000:80 & \
		kubectl port-forward -n storage svc/minio 9000:9000 & \
		kubectl port-forward -n live svc/live 8082:8082 & \
		kubectl port-forward -n monitoring svc/grafana 3001:3001 & \
		kubectl port-forward -n monitoring svc/victoria-logs 9428:9428 & \
		wait'


# 10.
apply-ingress:
	kubectl apply -f backend/k8s/ --recursive --selector stage=ingress



# Do not run this as it will probalby fail.
# Run all these in order but check that all db pods are complete and 
# running before running migrations
deploy-all: 
	$(MAKE) op-manifest
	$(MAKE) apply-kafka
	$(MAKE) apply-namespace
	$(MAKE) apply-pvc
	$(MAKE) apply-configs
	$(MAKE) apply-monitoring
	$(MAKE) apply-db1
	$(MAKE) apply-db2

#	Prod mode
# $(MAKE) deploy-nginx 

# 	wait for dbs
	sleep 60  
	$(MAKE) run-migrations 
	$(MAKE) apply-apps
	
# 	wait for storage
	sleep 10
	$(MAKE) apply-cors

# 	Dev mode
# 	wait for services
	sleep 30  
	$(MAKE) port-forward 

#	Prod mode
# 	$(MAKE) apply-ingress

# Runs the docker and k8s from top to bottom
first-time:
	$(MAKE) build-all
	$(MAKE) deploy-all
	

.PHONY: smart-full-start smart-apply-kafka smart-op-manifest smart-apply-namespace smart-apply-configs smart-apply-pvc smart-apply-monitoring smart-apply-db1 smart-apply-db2 smart-run-migrations smart-apply-apps smart-apply-cors smart-port-forward

big-mk:
	minikube start --cpus=10 --memory=12192m

define retry
	@for i in {1..100}; do \
		$(1) && break || (echo "Attempt $$i failed, retrying in 10s..."; sleep 5; echo -e "\n\n"); \
	done || (echo "All retries failed"; exit 1)
endef

smart-build-and-deploy:
	build-all
	$(MAKE) smart-full-deploy

smart-full-deploy:
# 	$(MAKE) load-images
# 	@sleep 2

	$(MAKE) smart-apply-kafka; 
	@sleep 2

	$(MAKE) smart-op-manifest; 
	@sleep 2

	$(MAKE) smart-apply-namespace; 
	@sleep 2

	$(MAKE) smart-apply-configs; 
	@sleep 2

	$(MAKE) smart-apply-pvc; 
	@sleep 2

	$(MAKE) smart-apply-monitoring; 
	@sleep 2

	$(call retry, $(MAKE) smart-apply-db1); 
	@sleep 2

	$(call retry, $(MAKE) smart-apply-db2); 
	@sleep 2
	@SLEEP=60; echo "Sleeping for $${SLEEP}s to wait for db's to start before migrations..."; echo -e "\n\n\n"; sleep $$SLEEP

	$(call retry, $(MAKE) smart-run-migrations); 
	@sleep 2
	@SLEEP=60; echo "Sleeping for $${SLEEP}s so that db's are ready before starting core services..."; echo -e "\n\n\n"; sleep $$SLEEP

	$(call retry, $(MAKE) smart-apply-apps); 
	@sleep 2

	$(call retry, $(MAKE) smart-apply-cors); 
	@sleep 2
	@SLEEP=100; echo "Sleeping for $${SLEEP}s to wait for pods to be ready before port forwarding them"; echo -e "\n\n\n"; sleep $$SLEEP

	$(MAKE) smart-port-forward

smart-apply-kafka:
	@echo " "
	@echo " "
	@echo " --- APPLYING KAFKA ---"
	@echo " "
	kubectl create namespace kafka || true
	$(call retry,kubectl apply -f "https://strimzi.io/install/latest?namespace=kafka" -n kafka)

smart-op-manifest:
	@echo " "
	@echo " "
	@echo " --- APPLYING CLOUDNATIVE-PG OPERATOR ---"
	@echo " "
	$(call retry,kubectl apply --server-side -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.28/releases/cnpg-1.28.0.yaml)

smart-apply-namespace:
	@echo " "
	@echo " "
	@echo " --- APPLYING NAMESPACES ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=namespace

smart-apply-configs:
	@echo " "
	@echo " "
	@echo " --- APPLYING CONFIGS ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=config

smart-apply-pvc:
	@echo " "
	@echo " "
	@echo " --- APPLYING PERSISTENT VOLUME CLAIMS ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=pvc

smart-apply-monitoring:
	@echo " "
	@echo " "
	@echo " --- APPLYING MONITORING ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=monitoring

smart-apply-db1:
	@echo " "
	@echo " "
	@echo " --- TRYING DATABASE PHASE ONE ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=db1

smart-apply-db2:
	@echo " "
	@echo " "
	@echo " --- TRYING DATABASE PHASE TWO ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=db2

smart-run-migrations:
	@echo " "
	@echo " "
	@echo " --- TRYING MIGRATIONS ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=migration

smart-apply-apps:
	@echo " "
	@echo " "
	@echo " --- TRYING TO START BUSINESS SERVICES ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=app

smart-apply-cors:
	@echo " "
	@echo " "
	@echo " --- TRYING TO APPLY CORS ---"
	@echo " "
	kubectl apply -R -f backend/k8s/ --selector stage=cors

smart-port-forward:
	@echo " "
	@echo " "
	@echo " --- STARTING PORT FORWARDS ---"
	@echo " "
	@bash -c ' \
	trap "pkill -f kubectl\ port-forward" EXIT SIGINT; \
	kubectl port-forward -n frontend svc/nextjs-frontend 3000:80 & \
	kubectl port-forward -n storage svc/minio 9000:9000 & \
	kubectl port-forward -n live svc/live 8082:8082 & \
	kubectl port-forward -n monitoring svc/grafana 3001:3001 & \
	kubectl port-forward -n monitoring svc/victoria-logs 9428:9428 & \
	wait'