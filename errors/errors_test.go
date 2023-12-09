package errors_test

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/ringsaturn/protoc-gen-go-errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestError struct{ message string }

func (e *TestError) Error() string { return e.message }

func TestErrors(t *testing.T) {
	var base *errors.Error
	err := errors.Newf(http.StatusBadRequest, "reason", "message")
	err2 := errors.Newf(http.StatusBadRequest, "reason", "message")
	err3 := err.WithMetadata(map[string]string{
		"foo": "bar",
	})
	werr := fmt.Errorf("wrap %w", err)

	if errors.Is(err, new(errors.Error)) {
		t.Errorf("should not be equal: %v", err)
	}
	if !errors.Is(werr, err) {
		t.Errorf("should be equal: %v", err)
	}
	if !errors.Is(werr, err2) {
		t.Errorf("should be equal: %v", err)
	}

	if !errors.As(err, &base) {
		t.Errorf("should be matches: %v", err)
	}
	if !errors.IsBadRequest(err) {
		t.Errorf("should be matches: %v", err)
	}

	if reason := errors.Reason(err); reason != err3.Reason {
		t.Errorf("got %s want: %s", reason, err)
	}

	if err3.Metadata["foo"] != "bar" {
		t.Error("not expected metadata")
	}

	gs := err.GRPCStatus()
	se := errors.FromError(gs.Err())
	if se.Reason != "reason" {
		t.Errorf("got %+v want %+v", se, err)
	}

	gs2 := status.New(codes.InvalidArgument, "bad request")
	se2 := errors.FromError(gs2.Err())
	// codes.InvalidArgument should convert to http.StatusBadRequest
	if se2.Code != http.StatusBadRequest {
		t.Errorf("convert code err, got %d want %d", errors.UnknownCode, http.StatusBadRequest)
	}
	if errors.FromError(nil) != nil {
		t.Errorf("FromError(nil) should be nil")
	}
	e := errors.FromError(stderrors.New("test"))
	if !reflect.DeepEqual(e.Code, int32(errors.UnknownCode)) {
		t.Errorf("no expect value: %v, but got: %v", e.Code, int32(errors.UnknownCode))
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name string
		e    *errors.Error
		err  error
		want bool
	}{
		{
			name: "true",
			e:    errors.New(404, "test", ""),
			err:  errors.New(http.StatusNotFound, "test", ""),
			want: true,
		},
		{
			name: "false",
			e:    errors.New(0, "test", ""),
			err:  stderrors.New("test"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok := tt.e.Is(tt.err); ok != tt.want {
				t.Errorf("Error.Error() = %v, want %v", ok, tt.want)
			}
		})
	}
}

func TestCause(t *testing.T) {
	testError := &TestError{message: "test"}
	err := errors.BadRequest("foo", "bar").WithCause(testError)
	if !errors.Is(err, testError) {
		t.Fatalf("want %v but got %v", testError, err)
	}
	if te := new(TestError); errors.As(err, &te) {
		if te.message != testError.message {
			t.Fatalf("want %s but got %s", testError.message, te.message)
		}
	}
}

func TestOther(t *testing.T) {
	err := errors.Errorf(10001, "test code 10001", "message")
	// Code
	if !reflect.DeepEqual(errors.Code(nil), 200) {
		t.Errorf("Code(nil) = %v, want %v", errors.Code(nil), 200)
	}
	if !reflect.DeepEqual(errors.Code(stderrors.New("test")), errors.UnknownCode) {
		t.Errorf(`Code(errors.New("test")) = %v, want %v`, errors.Code(nil), 200)
	}
	if !reflect.DeepEqual(errors.Code(err), 10001) {
		t.Errorf(`Code(err) = %v, want %v`, errors.Code(err), 10001)
	}
	// Reason
	if !reflect.DeepEqual(errors.Reason(nil), errors.UnknownReason) {
		t.Errorf(`Reason(nil) = %v, want %v`, errors.Reason(nil), errors.UnknownReason)
	}
	if !reflect.DeepEqual(errors.Reason(stderrors.New("test")), errors.UnknownReason) {
		t.Errorf(`Reason(errors.New("test")) = %v, want %v`, errors.Reason(nil), errors.UnknownReason)
	}
	if !reflect.DeepEqual(errors.Reason(err), "test code 10001") {
		t.Errorf(`Reason(err) = %v, want %v`, errors.Reason(err), "test code 10001")
	}
	// Clone
	err400 := errors.Newf(http.StatusBadRequest, "BAD_REQUEST", "param invalid")
	err400.Metadata = map[string]string{
		"key1": "val1",
		"key2": "val2",
	}
	if cerr := errors.Clone(err400); cerr == nil || cerr.Error() != err400.Error() {
		t.Errorf("Clone(err) = %v, want %v", errors.Clone(err400), err400)
	}
	if cerr := errors.Clone(nil); cerr != nil {
		t.Errorf("Clone(nil) = %v, want %v", errors.Clone(err400), err400)
	}
}
