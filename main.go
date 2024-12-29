package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/flum1025/sql-enum-generator/internal"
	"github.com/flum1025/sql-enum-generator/internal/app"
	"github.com/flum1025/sql-enum-generator/internal/entity"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	cmd := &cli.Command{
		Name:    "sql-enum-generator",
		Usage:   "Generate enum from SQL",
		Version: internal.Version,
		Commands: []*cli.Command{
			{
				Name: "generate",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "engine",
						Value: entity.EnginePostgres.String(),
						Action: func(_ context.Context, _ *cli.Command, value string) error {
							if _, err := entity.NewEngine(value); err != nil {
								return fmt.Errorf("invalid engine: %w", err)
							}

							return nil
						},
					},
					&cli.StringFlag{
						Name:  "config",
						Value: "sqlenumgen.yml",
					},
					&cli.StringFlag{
						Name:     "source-path",
						Usage:    "Wildcards can be used",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output-path",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					a, err := app.NewSchemaGenerator(app.SchemaGeneratorOption{
						Engine:     lo.Must(entity.NewEngine(cmd.String("engine"))),
						ConfigPath: cmd.String("config"),
						SourcePath: cmd.String("source-path"),
						OutputPath: cmd.String("output-path"),
					})
					if err != nil {
						return fmt.Errorf("new: %w", err)
					}

					if err := a.Run(); err != nil {
						return fmt.Errorf("run: %w", err)
					}

					return nil
				},
			},
			{
				Name: "query-generate",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "engine",
						Value: entity.EnginePostgres.String(),
						Action: func(_ context.Context, _ *cli.Command, value string) error {
							if _, err := entity.NewEngine(value); err != nil {
								return fmt.Errorf("invalid engine: %w", err)
							}

							return nil
						},
					},
					&cli.StringFlag{
						Name:  "config",
						Value: "sqlenumgen.yml",
					},
					&cli.StringFlag{
						Name:     "output-path",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "database-url",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					a, err := app.NewQueryGenerator(app.QueryGeneratorOption{
						Engine:      lo.Must(entity.NewEngine(cmd.String("engine"))),
						ConfigPath:  cmd.String("config"),
						OutputPath:  cmd.String("output-path"),
						DatabaseURL: cmd.String("database-url"),
					})
					if err != nil {
						return fmt.Errorf("new: %w", err)
					}

					if err := a.Run(); err != nil {
						return fmt.Errorf("run: %w", err)
					}

					return nil
				},
			},
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Fatalf("run: %v", err)
	}
}
