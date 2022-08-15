package main

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

// TODO: Refactor

const collectionName = "channels"

type record struct {
	Timestamp time.Time `firestore:"timestamp"`
	Amount    int64     `firestore:"amount"`
}

type storeClient struct {
	firestore *firestore.Client
}

func newStoreClient(ctx context.Context, projectID string) (*storeClient, error) {
	c, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, xerrors.Errorf("failed to build firestore client: %w", err)
	}
	return &storeClient{firestore: c}, nil
}

func (s *storeClient) collection(msg *slackMessage) *firestore.CollectionRef {
	return s.firestore.Collection(collectionName).Doc(msg.Event.Channel).Collection(msg.month())
}

func (s *storeClient) add(ctx context.Context, msg *slackMessage) error {
	docRef := s.collection(msg).Doc(msg.Event.TS)
	if _, err := docRef.Set(ctx, &record{msg.timestamp, msg.amount}); err != nil {
		return xerrors.Errorf("docRef.Set: %w", err)
	}

	return nil
}

func (s *storeClient) total(ctx context.Context, msg *slackMessage) (int64, error) {
	var r record
	var total int64

	docsIter := s.collection(msg).Documents(ctx)
	for {
		doc, err := docsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return -1, xerrors.Errorf("failed to get doc: %w", err)
		}
		if err := doc.DataTo(&r); err != nil {
			return -1, xerrors.Errorf("failed to map doc: %w", err)
		}
		total += r.Amount
	}

	return total, nil
}
