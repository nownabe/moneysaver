package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
)

const (
	firestoreEmulatorHost = "localhost:8937"
)

func TestMain(m *testing.M) {
	if err := os.Setenv("FIRESTORE_EMULATOR_HOST", firestoreEmulatorHost); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func test2Project(t *testing.T) string {
	t.Helper()

	name := t.Name()
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "")

	return strings.ToLower(name)
}

func getFirestoreClient(t *testing.T) *firestore.Client {
	t.Helper()

	fs, err := firestore.NewClient(context.Background(), test2Project(t))
	if err != nil {
		panic(err)
	}

	return fs
}

func flushStore(t *testing.T) {
	t.Helper()

	url := "http://" + firestoreEmulatorHost
	url += "/emulator/v1/projects/" + test2Project(t) + "/databases/(default)/documents"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode == 200 {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	panic(body)
}
