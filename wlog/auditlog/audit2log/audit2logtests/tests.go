// Copyright (c) 2018 Palantir Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package audit2logtests

import (
	"bytes"
	"io"
	"testing"

	"github.com/palantir/pkg/objmatcher"
	"github.com/palantir/pkg/safejson"
	"github.com/palantir/witchcraft-go-logging/wlog/auditlog/audit2log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	Name          string
	UID           string
	SID           string
	TokenID       string
	TraceID       string
	OtherUIDs     []string
	Origin        string
	AuditName     string
	AuditResult   audit2log.AuditResultType
	RequestParams map[string]interface{}
	ResultParams  map[string]interface{}
	JSONMatcher   objmatcher.MapMatcher
}

func (tc TestCase) Params() []audit2log.Param {
	return []audit2log.Param{
		audit2log.UID(tc.UID),
		audit2log.SID(tc.SID),
		audit2log.TokenID(tc.TokenID),
		audit2log.TraceID(tc.TraceID),
		audit2log.OtherUIDs(tc.OtherUIDs...),
		audit2log.Origin(tc.Origin),
		audit2log.RequestParams(tc.RequestParams),
		audit2log.ResultParams(tc.ResultParams),
	}
}

func TestCases() []TestCase {
	return []TestCase{
		{
			Name:          "basic audit log entry",
			UID:           "user-1",
			SID:           "session-1",
			TokenID:       "X-Y-Z",
			TraceID:       "trace-id-1",
			OtherUIDs:     []string{"user-2", "user-3"},
			Origin:        "0.0.0.0",
			AuditName:     "AUDITED_ACTION_NAME",
			AuditResult:   audit2log.AuditResultSuccess,
			RequestParams: map[string]interface{}{"requestKey": "requestValue"},
			ResultParams:  map[string]interface{}{"resultKey": "resultValue"},
			JSONMatcher: objmatcher.MapMatcher(map[string]objmatcher.Matcher{
				"time":      objmatcher.NewRegExpMatcher(".+"),
				"uid":       objmatcher.NewEqualsMatcher("user-1"),
				"sid":       objmatcher.NewEqualsMatcher("session-1"),
				"tokenId":   objmatcher.NewEqualsMatcher("X-Y-Z"),
				"traceId":   objmatcher.NewEqualsMatcher("trace-id-1"),
				"otherUids": objmatcher.NewEqualsMatcher([]interface{}{"user-2", "user-3"}),
				"origin":    objmatcher.NewEqualsMatcher("0.0.0.0"),
				"name":      objmatcher.NewEqualsMatcher("AUDITED_ACTION_NAME"),
				"result":    objmatcher.NewEqualsMatcher("SUCCESS"),
				"requestParams": objmatcher.MapMatcher(map[string]objmatcher.Matcher{
					"requestKey": objmatcher.NewEqualsMatcher("requestValue"),
				}),
				"resultParams": objmatcher.MapMatcher(map[string]objmatcher.Matcher{
					"resultKey": objmatcher.NewEqualsMatcher("resultValue"),
				}),
				"type": objmatcher.NewEqualsMatcher("audit.2"),
			}),
		},
	}
}

func JSONTestSuite(t *testing.T, loggerProvider func(w io.Writer) audit2log.Logger) {
	jsonOutputTests(t, loggerProvider)
	//jsonLoggerUpdateTest(t, loggerProvider)
}

func jsonOutputTests(t *testing.T, loggerProvider func(w io.Writer) audit2log.Logger) {
	for i, tc := range TestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := loggerProvider(buf)

			logger.Audit(
				tc.AuditName,
				tc.AuditResult,
				audit2log.UID(tc.UID),
				audit2log.SID(tc.SID),
				audit2log.TokenID(tc.TokenID),
				audit2log.TraceID(tc.TraceID),
				audit2log.OtherUIDs(tc.OtherUIDs...),
				audit2log.Origin(tc.Origin),
				audit2log.RequestParams(tc.RequestParams),
				audit2log.ResultParams(tc.ResultParams))

			gotAuditLog := map[string]interface{}{}
			logEntry := buf.Bytes()
			err := safejson.Unmarshal(logEntry, &gotAuditLog)
			require.NoError(t, err, "Case %d: %s\nAudit log line is not a valid map: %v", i, tc.Name, string(logEntry))

			assert.NoError(t, tc.JSONMatcher.Matches(gotAuditLog), "Case %d: %s", i, tc.Name)
		})
	}
}

//func jsonLoggerUpdateTest(t *testing.T, loggerProvider func(params wlog.LoggerParams, origin string) svc1log.Logger) {
//	t.Run("update JSON logger", func(t *testing.T) {
//		currCase := TestCases()[0]
//
//		buf := bytes.Buffer{}
//		logger := loggerProvider(wlog.LoggerParams{
//			Level:  wlog.ErrorLevel,
//			Output: &buf,
//		}, currCase.Origin)
//
//		// log at info level
//		logger.Info(currCase.Message, currCase.LogParams...)
//
//		// output should be empty
//		assert.Equal(t, "", buf.String())
//
//		// update configuration to log at info level
//		updatable, ok := logger.(wlog.UpdatableLogger)
//		require.True(t, ok, "logger does not support updating")
//
//		updated := updatable.UpdateLogger(wlog.LoggerParams{
//			Level:  wlog.InfoLevel,
//			Output: &buf,
//		})
//		assert.True(t, updated)
//
//		// log at info level
//		logger.Info(currCase.Message, currCase.LogParams...)
//
//		// output should exist and match
//		gotServiceLog := map[string]interface{}{}
//		logEntry := buf.Bytes()
//		err := safejson.Unmarshal(logEntry, &gotServiceLog)
//		require.NoError(t, err, "Service log line is not a valid map: %v", string(logEntry))
//
//		assert.NoError(t, currCase.JSONMatcher.Matches(gotServiceLog), "No match")
//	})
//}
