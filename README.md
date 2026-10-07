# sql-enum-generator

sql-enum-generator is a tool that converts SQL INSERT statements for master data into OpenAPI schemas. This application enables developers to easily generate enum representations of database master data using OpenAPI specifications, streamlining the development process and ensuring consistency between the database and application code.

Currently only postgresql is supported.

## Features

- Parses SQL INSERT statements and generates corresponding OpenAPI schemas
- Generates ID schemas whose type is resolved from the primary key column in CREATE TABLE statements
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

## ID Schemas

Add `id_type` to an entry of `tables` to generate an ID schema from the primary key of the table.

```yaml
version: "1"
tables:
  - name: products
    key: name
    value: id
    id_type: ProductID
  - name: menus
    key: name
    value: id
  - name: users
    id_type: UserID
```

| Key | Description |
| --- | --- |
| `name` | Table name. Schema-qualified and quoted names in DDL (e.g. `"public"."users"`) are matched by table name |
| `key`, `value` | Columns for the enum. Optional, but must be specified together. Without them, no enum is generated and master data is not required |
| `id_type` | ID schema name to generate. Optional |

Each entry requires `key` and `value`, or `id_type`.

When `id_type` is configured, pass the DDL files (CREATE TABLE statements) with `--schema-path`. Wildcards can be used. `--schema-path` is ignored when `id_type` is not configured.

```sh
$ go run github.com/flum1025/sql-enum-generator generate --source-path ./example/master.sql --schema-path ./example/schema.sql --output-path ./example/openapi.generated.json --config ./example/sqlenumgen.yml
```

The primary key must be a single column (table-level `PRIMARY KEY (...)` or column-level `PRIMARY KEY`), and its type is mapped as follows. Other types, composite primary keys, and missing tables result in an error.

| Column type | Schema |
| --- | --- |
| `uuid` | `{"type": "string", "format": "uuid"}` |
| `serial`, `integer` (`int`, `int4`) | `{"type": "integer", "format": "int32"}` |
| `bigserial`, `bigint` (`int8`) | `{"type": "integer", "format": "int64"}` |

Each ID schema has the `x-id: true` extension. When the entry also has `key` and `value`, the enum schema has the `x-id-type` extension that refers to the ID schema name.

```json
{
  "ProductID": { "type": "integer", "format": "int32", "x-id": true },
  "UserID": { "type": "string", "format": "uuid", "x-id": true },
  "products": {
    "enum": ["1", "2", "3"],
    "type": "integer",
    "x-enum-varnames": ["ProductA", "ProductB", "ProductC"],
    "x-id-type": "ProductID"
  }
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
  type-mapping:
    string:
      formats:
        uuid:
          type: string
```

`type-mapping` is optional. Without it, `format: uuid` is generated as `openapi_types.UUID` from `github.com/oapi-codegen/runtime`.

Then, run the following command to generate the Go code:

```sh
$ go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config ./example/oapi-codegen.yml ./example/openapi.generated.json
```

### TypeScript

For TypeScript, you can use [openapi-typescript](https://github.com/openapi-ts/openapi-typescript) to generate TypeScript definitions. Run the following command:

```sh
$ npx openapi-typescript ./example/openapi.generated.json -o ./example/openapi.generated.d.ts --enum
```

## Future Plans

- [ ] Add support for additional SQL dialects

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
