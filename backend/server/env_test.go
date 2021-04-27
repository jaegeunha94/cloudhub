package server

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	cloudhub "github.com/snetsystems/cloudhub/backend"
	"github.com/snetsystems/cloudhub/backend/log"
)

func TestEnvironment(t *testing.T) {
	type fields struct {
		Environment cloudhub.Environment
	}
	type wants struct {
		statusCode  int
		contentType string
		body        string
	}

	tests := []struct {
		name   string
		fields fields
		wants  wants
	}{
		{
			name: "Get environment",
			fields: fields{
				Environment: cloudhub.Environment{
					TelegrafSystemInterval: 1 * time.Minute,
				},
			},
			wants: wants{
				statusCode:  200,
				contentType: "application/json",
				body:        `{"links":{"self":"/cloudhub/v1/env"},"telegrafSystemInterval":"1m0s"}`,
			},
		},
		{
			name: "Get environment with CustomAutoRefresh",
			fields: fields{
				Environment: cloudhub.Environment{
					TelegrafSystemInterval: 2 * time.Minute,
					CustomAutoRefresh:      "500ms=500",
				},
			},
			wants: wants{
				statusCode:  200,
				contentType: "application/json",
				body:        `{"links":{"self":"/cloudhub/v1/env"},"telegrafSystemInterval":"2m0s","customAutoRefresh": "500ms=500"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				Env:    tt.fields.Environment,
				Logger: log.New(log.DebugLevel),
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://any.url", nil)

			s.Environment(w, r)

			resp := w.Result()
			content := resp.Header.Get("Content-Type")
			body, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode != tt.wants.statusCode {
				t.Errorf("%q. Config() = %v, want %v", tt.name, resp.StatusCode, tt.wants.statusCode)
			}
			if tt.wants.contentType != "" && content != tt.wants.contentType {
				t.Errorf("%q. Config() = %v, want %v", tt.name, content, tt.wants.contentType)
			}
			if eq, _ := jsonEqual(string(body), tt.wants.body); tt.wants.body != "" && !eq {
				t.Errorf("%q. Config() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wants.body)
			}
		})
	}
}
