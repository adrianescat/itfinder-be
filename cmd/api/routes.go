package main

import (
	"database/sql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/julienschmidt/httprouter"
	"itfinder.adrianescat.com/graph"
	"itfinder.adrianescat.com/graph/model"
	"net/http"
)

func (app *app) routes(db *sql.DB) http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	gql := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		Models: model.NewModels(db),
		Logger: app.logger,
		Wg:     app.wg,
	}}))

	plg := playground.Handler("GraphQL playground", "/query")

	router.Handle(http.MethodPost, "/query", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		gql.ServeHTTP(w, req)
	})

	router.Handle(http.MethodGet, "/", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		plg.ServeHTTP(w, req)
	})

	return router
}
