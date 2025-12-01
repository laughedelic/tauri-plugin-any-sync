package dispatcher

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestDispatcher_Register(t *testing.T) {
	d := New()

	handler := func(ctx context.Context, req proto.Message) (proto.Message, error) {
		return &emptypb.Empty{}, nil
	}

	d.Register("test", handler, &emptypb.Empty{})

	commands := d.Commands()
	if len(commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(commands))
	}
	if commands[0] != "test" {
		t.Errorf("expected command 'test', got '%s'", commands[0])
	}
}

func TestDispatcher_Dispatch_Success(t *testing.T) {
	d := New()

	handler := func(ctx context.Context, req proto.Message) (proto.Message, error) {
		// Echo the request value in response
		stringReq := req.(*wrapperspb.StringValue)
		return wrapperspb.String(stringReq.Value + "_response"), nil
	}

	d.Register("echo", handler, &wrapperspb.StringValue{})

	req := wrapperspb.String("hello")
	reqBytes, _ := proto.Marshal(req)

	respBytes, err := d.Dispatch(context.Background(), "echo", reqBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp wrapperspb.StringValue
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	expected := "hello_response"
	if resp.Value != expected {
		t.Errorf("expected '%s', got '%s'", expected, resp.Value)
	}
}

func TestDispatcher_Dispatch_UnknownCommand(t *testing.T) {
	d := New()

	_, err := d.Dispatch(context.Background(), "unknown", []byte{})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}

	expectedErr := "unknown command: unknown"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestDispatcher_Dispatch_InvalidPayload(t *testing.T) {
	d := New()

	handler := func(ctx context.Context, req proto.Message) (proto.Message, error) {
		return &emptypb.Empty{}, nil
	}

	d.Register("test", handler, &emptypb.Empty{})

	// Invalid protobuf payload
	_, err := d.Dispatch(context.Background(), "test", []byte{0xFF, 0xFF})
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestDispatcher_Dispatch_HandlerError(t *testing.T) {
	d := New()

	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, req proto.Message) (proto.Message, error) {
		return nil, expectedErr
	}

	d.Register("test", handler, &emptypb.Empty{})

	req := &emptypb.Empty{}
	reqBytes, _ := proto.Marshal(req)

	_, err := d.Dispatch(context.Background(), "test", reqBytes)
	if err == nil {
		t.Fatal("expected handler error")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}
