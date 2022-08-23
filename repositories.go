package main

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errNotFound = errors.New("not found")

type channelRepo struct {
	*firestore.Client
}

func (r *channelRepo) findByID(ctx context.Context, chID string) (*channel, error) {
	docRef := r.Collection(collectionName).Doc(chID)

	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errNotFound
		}

		return nil, fmt.Errorf("r.Collection.Doc: %w", err)
	}

	var ch channel
	if err := doc.DataTo(&ch); err != nil {
		return nil, fmt.Errorf("doc.DataTo: %w", err)
	}

	return &ch, nil
}

func (r *channelRepo) save(ctx context.Context, ch *channel) error {
	docRef := r.Collection(collectionName).Doc(ch.ID)
	if _, err := docRef.Set(ctx, ch); err != nil {
		return fmt.Errorf("docRef.Set: %w", err)
	}

	return nil
}
