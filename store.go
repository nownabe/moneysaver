package main

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

const collectionName = "channels"

type record struct {
	Timestamp time.Time `firestore:"timestamp"`
	Amount    int64     `firestore:"amount"`
}

func addRecord(ctx context.Context, msg *slackMessage) error {
	client, err := firestore.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		return xerrors.Errorf("failed to build firestore client: %w", err)
	}

	chDoc := client.Collection(collectionName).Doc(msg.Event.Channel)

	if _, _, err := chDoc.Collection(msg.month()).Add(ctx, &record{
		Timestamp: msg.timestamp,
		Amount:    msg.amount,
	}); err != nil {
		return xerrors.Errorf("failed to add document: %w", err)
	}

	return nil
}

func aggregate(ctx context.Context, msg *slackMessage) (int64, error) {
	client, err := firestore.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		return -1, xerrors.Errorf("failed to build firestore client: %w", err)
	}

	chDoc := client.Collection(collectionName).Doc(msg.Event.Channel)

	var r record
	var total int64

	docsIter := chDoc.Collection(msg.month()).Documents(ctx)
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
