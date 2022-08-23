package main

// https://firebase.google.com/docs/firestore/manage-data/data-types?hl=ja#data_types
type channel struct {
	ID     string `firestore:"-"`
	Budget int64  `firestore:"budget"`
}

type expenditure struct {
	ts     string // Slack timestamp which is used to identify the message
	amount int64
}
