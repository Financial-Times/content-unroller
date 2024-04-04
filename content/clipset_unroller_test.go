package content

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUnroller struct {
	unrollFunc func(event UnrollEvent) (Content, error)
}

func (m mockUnroller) Unroll(event UnrollEvent) (Content, error) {
	return m.unrollFunc(event)
}

func TestClipsetUnroller_Unroll(t *testing.T) {
	defaultAPIHost := "test.api.ft.com"
	testTID := "testTID"
	testUUID := "testUUID"
	testUUIDClip := "22c0d426-1466-11e7-b0c1-37e417ee6c76"
	invalidClipset := Content{
		typeField: "wrong",
	}
	validClipsetWithNoMembers := Content{
		membersField: []interface{}{},
		typeField:    ClipSetType,
	}
	unrolledClip := Content{
		id:         testUUIDClip,
		"unrolled": "true",
	}
	type fields struct {
		clipUnroller Unroller
		reader       Reader
		apiHost      string
	}
	tests := []struct {
		name           string
		unrollerFields fields
		event          UnrollEvent
		want           Content
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name: "invalid-clipset",
			unrollerFields: fields{
				clipUnroller: nil,
				reader:       nil,
				apiHost:      defaultAPIHost,
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
				clipUnroller: nil,
				reader:       nil,
				apiHost:      defaultAPIHost,
			},
			event: UnrollEvent{
				c:    validClipsetWithNoMembers,
				tid:  testTID,
				uuid: testUUID,
			},
			want:    validClipsetWithNoMembers,
			wantErr: assert.NoError,

			//want:    nil,
			//wantErr: assert.Error, //TODO: Do we want error when sending empty members?
		},
		{
			name: "valid-clipset-with-members",
			unrollerFields: fields{
				clipUnroller: mockUnroller{
					unrollFunc: func(event UnrollEvent) (Content, error) {
						return event.c, nil
					},
				},
				reader: &ReaderMock{
					mockGet: func(uuids []string, tid string) (map[string]Content, error) {
						return map[string]Content{
							testUUIDClip: unrolledClip,
						}, nil
					},
				},
				apiHost: defaultAPIHost,
			},
			event: UnrollEvent{
				c: Content{
					membersField: []interface{}{
						map[string]interface{}{
							id: testUUIDClip,
						},
					},
					typeField: ClipSetType,
				},
				tid:  testTID,
				uuid: testUUID,
			},
			want: Content{
				membersField: []Content{
					unrolledClip,
				},
				typeField: ClipSetType,
			},
			wantErr: assert.NoError,
		},
		//TODO: Test with real clip validator maybe?
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &ClipsetUnroller{
				clipUnroller: tt.unrollerFields.clipUnroller,
				reader:       tt.unrollerFields.reader,
				apiHost:      tt.unrollerFields.apiHost,
			}
			got, err := u.Unroll(tt.event)
			if !tt.wantErr(t, err, fmt.Sprintf("Unroll(%v)", tt.event)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Unroll(%v)", tt.event)
		})
	}
}

func Test_validateClipset(t *testing.T) {
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
				typeField: ClipSetType,
			},
			want: false,
		},
		{
			name: "content-with-correct-type-and-members",
			content: Content{
				typeField:    ClipSetType,
				membersField: []Content{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateClipset(tt.content), "validateClipset(%v)", tt.content)
		})
	}
}
