build:
	@docker build -t cloudtrail-query:latest .

shell: build
	@docker run -it --rm -v ~/.aws:/root/.aws --entrypoint sh cloudtrail-query:latest