include .env.dev
export

.PHONY: build
build:
	docker build -t ziliscite/video-to-mp4-gateway:latest .

.PHONY: push
push:
	docker push ziliscite/video-to-mp4-gateway:latest

.PHONY: deploy
deploy:
	kubectl apply -f ./
