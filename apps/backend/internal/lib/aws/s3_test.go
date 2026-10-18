package aws

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/config"
)

func TestCreatePresignedUploadURLValidation(t *testing.T) {
	s := NewStorage(NewS3Client(aws.Config{Region: "us-east-1"}, config.S3Config{}), "b")
	ctx := context.Background()

	if _, _, err := s.CreatePresignedUploadURL(ctx, "", "", 10, time.Minute); err == nil {
		t.Error("expected error for empty content type")
	}
	if _, _, err := s.CreatePresignedUploadURL(ctx, "", "image/png", 0, time.Minute); err == nil {
		t.Error("expected error for size 0")
	}
}

func TestKeyOrUUID(t *testing.T) {
	if got := keyOrUUID("avatars/1.png"); got != "avatars/1.png" {
		t.Errorf("keyOrUUID kept key = %q", got)
	}
	if a, b := keyOrUUID(""), keyOrUUID(""); a == "" || a == b {
		t.Errorf("keyOrUUID(\"\") = %q, %q; want distinct UUIDs", a, b)
	}
}

// The tests below run against the docker-compose RustFS:
//
//	S3_TEST_ENDPOINT=http://localhost:9000 go test ./internal/lib/aws
func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	endpoint := os.Getenv("S3_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("S3_TEST_ENDPOINT not set")
	}
	a, err := NewAWS(context.Background(), config.AWSConfig{
		Region:          "us-east-1",
		AccessKeyID:     "rustfsadmin",
		SecretAccessKey: "rustfsadmin",
		S3:              config.S3Config{EndpointURL: endpoint, Bucket: "boilerplate"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return a.S3
}

func TestNewAWSMissingBucket(t *testing.T) {
	endpoint := os.Getenv("S3_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("S3_TEST_ENDPOINT not set")
	}
	_, err := NewAWS(context.Background(), config.AWSConfig{
		Region:          "us-east-1",
		AccessKeyID:     "rustfsadmin",
		SecretAccessKey: "rustfsadmin",
		S3:              config.S3Config{EndpointURL: endpoint, Bucket: "does-not-exist"},
	})
	if err == nil {
		t.Fatal("expected startup error for missing bucket")
	}
}

func get(t *testing.T, url string) (int, string, string) {
	t.Helper()
	resp, err := http.Get(url) //nolint:gosec,noctx // test URL from our own presigner
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(body), resp.Header.Get("Content-Type")
}

func put(t *testing.T, url, contentType, body string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body)) //nolint:noctx // test only
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestStorageRoundTrip(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	key, err := s.Upload(ctx, "", strings.NewReader("hello")) // shorter than the 512-byte sniff
	if err != nil {
		t.Fatal(err)
	}

	url, err := s.CreatePresignedURL(ctx, key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	code, body, ct := get(t, url)
	if code != http.StatusOK || body != "hello" || !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("get: status=%d body=%q type=%q", code, body, ct)
	}

	if err := s.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := get(t, url); code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", code)
	}
}

func TestPresignedUpload(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()
	const body = "fake-png"

	url, key, err := s.CreatePresignedUploadURL(ctx, "test/avatar.png", "image/png", int64(len(body)), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if key != "test/avatar.png" {
		t.Fatalf("key = %q, want caller-provided key", key)
	}
	t.Cleanup(func() { _ = s.Delete(ctx, key) })

	if code := put(t, url, "image/jpeg", body); code != http.StatusForbidden {
		t.Fatalf("wrong content type: status=%d, want 403", code)
	}
	if code := put(t, url, "image/png", body+"x"); code != http.StatusForbidden {
		t.Fatalf("wrong size: status=%d, want 403", code)
	}
	if code := put(t, url, "image/png", body); code != http.StatusOK {
		t.Fatalf("valid upload: status=%d, want 200", code)
	}

	getURL, err := s.CreatePresignedURL(ctx, key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if code, got, ct := get(t, getURL); code != http.StatusOK || got != body || ct != "image/png" {
		t.Fatalf("get: status=%d body=%q type=%q", code, got, ct)
	}
}
