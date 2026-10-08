# sql-enum-generator

sql-enum-generator is a tool that converts SQL INSERT statements for master data into OpenAPI schemas. This application enables developers to easily generate enum representations of database master data using OpenAPI specifications, streamlining the development process and ensuring consistency between the database and application code.

Currently only postgresql is supported.

## Features

- Parses SQL INSERT statements and generates corresponding OpenAPI schemas
- Generated OpenAPI schemas can be utilized with other tools for type generation

## Quick Start

1. **Create a configuration file** named `sqlenumgen.yml` with the following content:

```yaml
version: "1"
tables:
  - name: products
    key: name
    value: id
```

2. **Run the application** using the following command:

```sh
$ go run github.com/flum1025/sql-enum-generator generate --source-path ./example/master.sql --output-path ./example/openapi.generated.json --config ./example/sqlenumgen.yml
```

3. **Utilize language-specific generation tools** to create enums from the generated OpenAPI schema.

For actual generation examples, please refer to the `example` directory in the repository.

## ID Type

Add optional `id_type` to a table to add the `x-id-type` extension to its enum schema.

```yaml
version: "1"
tables:
  - name: products
    key: name
    value: id
    id_type: ProductID
```

```json
"products": {
  "enum": ["1", "2", "3"],
  "type": "integer",
  "x-enum-varnames": ["ProductA", "ProductB", "ProductC"],
  "x-id-type": "ProductID"
}
```

`x-id-type` is not used by the default templates of oapi-codegen or by openapi-typescript. It allows custom templates to generate conversions from an enum to the ID type defined in your code. For example, with oapi-codegen `user-templates`, appending the following to `constants.tmpl`:

```
{{range $Enum := .EnumDefinitions}}
{{- with index $Enum.Schema.OAPISchema.Extensions "x-id-type"}}
func (v {{$Enum.TypeName}}) {{.}}() {{.}} {
	return {{.}}(v)
}
{{end}}
{{- end}}
```

generates:

```go
func (v Products) ProductID() ProductID {
	return ProductID(v)
}
```

## Language-Specific Usage Examples

### Go

For Go, you can use [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to generate code from the OpenAPI schema. Create a configuration file named `oapi-codegen.yml` with the following content:

```yaml
package: main
output: ./openapi.generated.go
generate:
  models: true
compatibility:
  always-prefix-enum-values: true
output-options:
  skip-prune: true
```

Then, run the following command to generate the Go code:

```sh
$ go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config ./example/oapi-codegen.yml ./example/openapi.generated.json
```

### TypeScript

For TypeScript, you can use [openapi-typescript](https://github.com/openapi-ts/openapi-typescript) to generate TypeScript definitions. Run the following command:

```sh
$ npx openapi-typescript ./example/openapi.generated.json -o ./example/openapi.generated.d.ts --enum
```

## Development

```sh
$ bun install
```

`bun install` sets up a husky `pre-push` hook that regenerates `example/` with `make example` and fails if the generated files differ from the committed ones, then type-checks `example/openapi.generated.d.ts` and runs `go vet ./...` and `go test ./...`.

## Future Plans

- [ ] Add support for additional SQL dialects

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
