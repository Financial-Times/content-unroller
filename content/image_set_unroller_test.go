package content

import (
	"fmt"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestUniversalUnroller_unrollImageSet(t *testing.T) {
	testLogger := logger.NewUPPLogger("test-service", "Error")
	defaultAPIHost := "test.api.ft.com"
	testTID := "testTID"
	testUUID := "testUUID"
	testUUIDClip := "22c0d426-1466-11e7-b0c1-37e417ee6c76"
	invalidClipset := Content{
		typeField: "wrong",
	}
	validClipsetWithNoMembers := Content{
		membersField: []interface{}{},
		typeField:    ImageSetType,
	}
	unrolledClip := Content{
		id:         testUUIDClip,
		"unrolled": "true",
		typeField:  image,
	}
	type fields struct {
		reader  Reader
		log     *logger.UPPLogger
		apiHost string
	}
	tests := []struct {
		name           string
		unrollerFields fields
		event          UnrollEvent
		want           Content
		wantErr        assert.ErrorAssertionFunc
	}{{
		name: "invalid-clipset",
		unrollerFields: fields{
			reader:  nil,
			log:     testLogger,
			apiHost: defaultAPIHost,
		},
		event: UnrollEvent{
			c:    invalidClipset,
			tid:  testTID,
			uuid: testUUID,
		},
		want:    nil,
		wantErr: assert.Error,
	},
		{
			name: "valid-clipset-with-no-members",
			unrollerFields: fields{
				reader:  nil,
				log:     testLogger,
				apiHost: defaultAPIHost,
			},
			event: UnrollEvent{
				c:    validClipsetWithNoMembers,
				tid:  testTID,
				uuid: testUUID,
			},
			want:    validClipsetWithNoMembers,
			wantErr: assert.NoError,
		},
		{
			name: "valid-clipset-with-members",
			unrollerFields: fields{
				reader: &ReaderMock{
					mockGet: func(_ []string, _ string) (map[string]Content, error) {
						return map[string]Content{
							testUUIDClip: unrolledClip,
						}, nil
					},
				},
				log:     testLogger,
				apiHost: defaultAPIHost,
			},
			event: UnrollEvent{
				c: Content{
					membersField: []interface{}{
						map[string]interface{}{
							id: testUUIDClip,
						},
					},
					typeField: ImageSetType,
				},
				tid:  testTID,
				uuid: testUUID,
			},
			want: Content{
				membersField: []Content{
					unrolledClip,
				},
				typeField: ImageSetType,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UniversalUnroller{
				reader:  tt.unrollerFields.reader,
				log:     tt.unrollerFields.log,
				apiHost: tt.unrollerFields.apiHost,
			}
			got, err := u.unrollImageSet(tt.event)
			if !tt.wantErr(t, err, fmt.Sprintf("unrollImageSet(%v)", tt.event)) {
				return
			}
			assert.Equalf(t, tt.want, got, "unrollImageSet(%v)", tt.event)
		})
	}
}

func Test_validateImageSet(t *testing.T) {
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
			name: "content-with-correct-type-and-no-members",
			content: Content{
				typeField: ImageSetType,
			},
			want: false,
		},
		{
			name: "content-with-correct-type-and-members",
			content: Content{
				typeField:    ImageSetType,
				membersField: []Content{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateImageSet(tt.content), "validateImageSet(%v)", tt.content)
		})
	}
}
