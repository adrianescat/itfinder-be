package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.24

import (
	"context"
	"fmt"
	"itfinder.adrianescat.com/graph/model"
)

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUserInput) (*model.User, error) {
	user := &model.User{
		Name:      input.Name,
		Lastname:  input.LastName,
		Email:     input.Email,
		Activated: false,
	}

	err := r.Models.Users.Insert(user)

	if err != nil {
		r.Logger.PrintError(fmt.Errorf("%s", err), nil)
		return nil, err
	}

	return user, nil

}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	panic(fmt.Errorf("not implemented: Users - users"))
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
