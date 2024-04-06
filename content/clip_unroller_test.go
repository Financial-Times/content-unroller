package content

import (
	"fmt"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestClipUnroller_Unroll(t *testing.T) {
	testLogger := logger.NewUPPLogger("test-service", "Error")
	type fields struct {
		imageSetUnroller Unroller
		reader           Reader
		apiHost          string
	}
	tests := []struct {
		name    string
		fields  fields
		event   UnrollEvent
		want    Content
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUniversalUnroller(tt.fields.reader, testLogger, tt.fields.apiHost)
			got, err := u.UnrollContent(tt.event)
			if !tt.wantErr(t, err, fmt.Sprintf("Unroll(%v)", tt.event)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Unroll(%v)", tt.event)
		})
	}
}

func Test_validateClip(t *testing.T) {
	tests := []struct {
		name    string
		content Content
		want    bool
	}{
		{
			name:    "empty",
			content: Content{},
			want:    false,
		},
		{
			name: "content-with-wrong-type",
			content: Content{
				typeField: "wrong",
			},
			want: false,
		},
		{
			name: "content-with-correct-type",
			content: Content{
				typeField: ClipType,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateClip(tt.content), "validateClip(%v)", tt.content)
		})
	}
}
