SHELL=/bin/bash

.PHONY: dummy
dummy:
	@echo argument is required

.PHONY: example
example:
	go run github.com/flum1025/sql-enum-generator generate --source-path ./example/master.sql --output-path ./example/openapi.generated.json --config ./example/sqlenumgen.yml
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config ./example/oapi-codegen.yml ./example/openapi.generated.json
	npx openapi-typescript ./example/openapi.generated.json -o ./example/openapi.generated.d.ts --enum
