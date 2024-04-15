package content

import (
	"fmt"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestClipUnroller_Unroll(t *testing.T) {
	testLogger := logger.NewUPPLogger("test-service", "Error")
	testAPIHost := "test.api.ft.com"
	posterUUID := "22c0d426-1466-11e7-b0c1-37e417ee6c76"
	type fields struct {
		reader  Reader
		apiHost string
	}
	imageUUID := "22c0d426-1466-11e7-b0c1-37e417ee6c77"
	tests := []struct {
		name    string
		fields  fields
		event   UnrollEvent
		want    Content
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "invalid-clip",
			fields: fields{
				reader:  nil,
				apiHost: testAPIHost,
			},
			event: UnrollEvent{
				c: Content{
					typeField: "wrong",
				},
			},
			want: Content{
				typeField: "wrong",
			},
			wantErr: assert.Error,
		},
		{
			name: "valid-clip-without-poster",
			fields: fields{
				reader:  nil,
				apiHost: testAPIHost,
			},
			event: UnrollEvent{
				c: Content{
					typeField: ClipType,
				},
			},
			want: Content{
				typeField: ClipType,
			},
			wantErr: assert.NoError,
		},
		{
			name: "valid-clip-with-poster",
			fields: fields{
				reader: &ReaderMock{
					mockGet: func(_ []string, _ string) (map[string]Content, error) {
						return map[string]Content{
							posterUUID: {
								id:         posterUUID,
								"unrolled": "true",
								typeField:  ImageSetType,
								membersField: []interface{}{
									map[string]interface{}{
										id: imageUUID,
									},
								},
							},
							imageUUID: {
								id:         imageUUID,
								typeField:  image,
								"unrolled": "true",
							},
						}, nil
					},
				},
				apiHost: testAPIHost,
			},
			event: UnrollEvent{
				c: Content{
					posterField: map[string]interface{}{
						apiURLField: posterUUID,
					},
					typeField: ClipType,
				},
			},
			want: Content{
				posterField: Content{
					id:         posterUUID,
					"unrolled": "true",
					typeField:  ImageSetType,
					membersField: []Content{
						{
							id:         imageUUID,
							typeField:  image,
							"unrolled": "true",
						},
					},
				},
				typeField: ClipType,
			},
			wantErr: assert.NoError,
		},
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
