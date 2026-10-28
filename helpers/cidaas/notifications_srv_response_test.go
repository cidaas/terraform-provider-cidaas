package cidaas

import (
	"net/http"
	"testing"
)

func TestParseNotificationSrvGraphList(t *testing.T) {
	t.Parallel()
	type row struct {
		ID string `json:"_id"`
	}
	tests := []struct {
		name    string
		status  int
		body    string
		want    int
		wantErr bool
	}{
		{name: "204 no content is an empty result", status: http.StatusNoContent, body: "", want: 0},
		{name: "200 with rows", status: http.StatusOK, body: `{"success":true,"status":200,"data":[{"_id":"a"},{"_id":"b"}]}`, want: 2},
		{name: "200 with empty array", status: http.StatusOK, body: `{"success":true,"status":200,"data":[]}`, want: 0},
		{name: "error envelope", status: http.StatusBadRequest, body: `{"success":false,"status":400,"errorMsg":"invalid filter"}`, wantErr: true},
		{name: "empty body on error status", status: http.StatusInternalServerError, body: "", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseNotificationSrvGraphList[row]([]byte(tc.body), tc.status)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tc.want {
				t.Fatalf("expected %d rows, got %d", tc.want, len(got))
			}
		})
	}
}
