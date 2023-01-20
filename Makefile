generate-api:
	java -jar ~/openapi-generator-cli.jar generate -i api/openapi.yaml -g go-gin-server -o ./pkg/api --skip-validate-spec
