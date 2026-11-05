// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAliasEmptyContent(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		want          int
	}{
		{"sentinel", `{"":{"value":"","selected":1}}`, 0},
		{"scalar", `""`, 0},
		{"map", `{}`, 0},
		{"array", `[]`, 0},
		{"mixed", `{"":{"value":"","selected":1},"192.0.2.1":{"value":"192.0.2.1","selected":1}}`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var api aliasAPIResponse
			if err := json.Unmarshal([]byte(`{"content":`+tc.content+`}`), &api); err != nil {
				t.Fatal(err)
			}
			var model AliasResourceModel
			model.fromAPI(context.Background(), &api, "fixture")
			if model.Content.IsNull() || len(model.Content.Elements()) != tc.want {
				t.Fatalf("unexpected content: %v", model.Content)
			}
			req := model.toAPI(context.Background())
			if tc.want == 0 && req.Content != "" {
				t.Fatalf("empty content serialized as %q", req.Content)
			}
			if tc.want == 1 && req.Content != "192.0.2.1" {
				t.Fatalf("nonempty content lost: %q", req.Content)
			}
		})
	}
}
