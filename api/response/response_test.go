package response

import "testing"

func TestResponse(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		resp := Success()
		if resp.Success != true {
			t.Error("expected true success in response")
		}
	})

	t.Run("error response", func(t *testing.T) {
		resp := NewError("TEST", "test message")
		if resp.Success != false {
			t.Error("expected false success in response")
		}
		if resp.Error.Code != "TEST" {
			t.Errorf("expected \"TEST\" error code got %q", resp.Error.Code)
		}
		if resp.Error.Message != "test message" {
			t.Errorf("expected \"test message\" error message got %q", resp.Error.Message)
		}
	})
}
