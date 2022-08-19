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

func (s *storeClient) collection(channel string, ts time.Time) *firestore.CollectionRef {
	month := ts.Format("2006-01")
	return s.firestore.Collection(collectionName).Doc(channel).Collection(month)
}

func (s *storeClient) add(ctx context.Context, channel string, ts time.Time, ex *expenditure) error {
	docRef := s.collection(channel, ts).Doc(ex.ts)
	if _, err := docRef.Set(ctx, &record{ts, ex.amount}); err != nil {
		return xerrors.Errorf("docRef.Set: %w", err)
	}

	return nil
}

func (s *storeClient) total(ctx context.Context, channel string, ts time.Time) (int64, error) {
	var r record
	var total int64

	docsIter := s.collection(channel, ts).Documents(ctx)
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
