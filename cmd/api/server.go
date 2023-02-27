package main

import (
	"database/sql"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"itfinder.adrianescat.com/graph"
	"itfinder.adrianescat.com/graph/model"
	"net/http"
)

type ApplicationKeyType string

var appKey ApplicationKeyType = "APP"

func (app *app) serve(db *sql.DB) error {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		Models: model.NewModels(db),
		Logger: app.logger,
		Wg:     app.wg,
	}}))

	// TODO: shutdown

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": fmt.Sprintf(":%d", app.config.port),
		"env":  app.config.env,
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	app.logger.PrintInfo("connect for GraphQL playground", map[string]string{
		"addr": fmt.Sprintf(":%d/", app.config.port),
	})

	err := http.ListenAndServe(fmt.Sprintf(":%d", app.config.port), nil)

	if err != nil {
		panic(err)
	}

	return nil
}
