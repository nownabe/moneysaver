package main

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const collectionName = "channels"

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

type expenditureRepo struct {
	*firestore.Client
}

func (r *expenditureRepo) collection(ex *expenditure) *firestore.CollectionRef {
	month := ex.Timestamp.Format("2006-01")
	return r.Collection(collectionName).Doc(ex.Channel).Collection(month)
}

func (r *expenditureRepo) add(ctx context.Context, ex *expenditure) error {
	docRef := r.collection(ex).Doc(ex.TS)
	if _, err := docRef.Set(ctx, ex); err != nil {
		return fmt.Errorf("docRef.Set: %w", err)
	}

	return nil
}

func (r *expenditureRepo) total(ctx context.Context, ex *expenditure) (int64, error) {
	var e *expenditure
	var total int64

	docsIter := r.collection(ex).Documents(ctx)
	for {
		doc, err := docsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return -1, fmt.Errorf("docsIter.Next: %w", err)
		}
		if err := doc.DataTo(&e); err != nil {
			return -1, fmt.Errorf("doc.DataTo: %w", err)
		}
		total += e.Amount
	}

	return total, nil
}
