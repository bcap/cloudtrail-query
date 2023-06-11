build:
	@docker build -t bcap/cloudtrail-query:latest .

shell: build
	@docker run -it --rm -v ~/.aws:/root/.aws --entrypoint sh bcap/cloudtrail-query:latest

push: build
	@docker push bcap/cloudtrail-query:latest

pull:
	@docker pull bcap/cloudtrail-query:latest
