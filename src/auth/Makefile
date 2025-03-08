include .env.dev
export

.PHONY: build
build:
	docker build -t ziliscite/video-to-mp4-auth:latest .

.PHONY: push
push:
	docker push ziliscite/video-to-mp4-auth:latest

.PHONY: run
run:
	docker run -p 5000:80 \
	-e DB_DSN=$(DB_DSN) \
	-e JWT_SECRETS=$(JWT_SECRETS) \
	ziliscite/video-to-mp4-auth

.PHONY: deploy
deploy:
	kubectl apply -f ./


