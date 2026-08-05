/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"testing"
	"time"
)

func TestResolveOperationTimeouts(t *testing.T) {
	tests := []struct {
		name          string
		explicitFlags []string
		timeout       time.Duration
		resizeTimeout time.Duration
		modifyTimeout time.Duration
		wantResize    time.Duration
		wantModify    time.Duration
	}{
		{
			name:          "nothing explicit, defaults used",
			timeout:       10 * time.Second,
			resizeTimeout: 10 * time.Second,
			modifyTimeout: 10 * time.Second,
			wantResize:    10 * time.Second,
			wantModify:    10 * time.Second,
		},
		{
			name:          "only resize-timeout and modify-timeout set explicitly",
			explicitFlags: []string{"resize-timeout", "modify-timeout"},
			timeout:       10 * time.Second,
			resizeTimeout: 300 * time.Second,
			modifyTimeout: 1800 * time.Second,
			wantResize:    300 * time.Second,
			wantModify:    1800 * time.Second,
		},
		{
			name:          "timeout set explicitly overrides both, even with resize/modify also set",
			explicitFlags: []string{"timeout", "resize-timeout", "modify-timeout"},
			timeout:       120 * time.Second,
			resizeTimeout: 300 * time.Second,
			modifyTimeout: 1800 * time.Second,
			wantResize:    120 * time.Second,
			wantModify:    120 * time.Second,
		},
		{
			name:          "timeout set explicitly at its own default value still overrides",
			explicitFlags: []string{"timeout"},
			timeout:       10 * time.Second,
			resizeTimeout: 300 * time.Second,
			modifyTimeout: 1800 * time.Second,
			wantResize:    10 * time.Second,
			wantModify:    10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet(tt.name, flag.ContinueOnError)
			fs.Duration("timeout", 10*time.Second, "")
			fs.Duration("resize-timeout", 10*time.Second, "")
			fs.Duration("modify-timeout", 10*time.Second, "")

			var args []string
			for _, name := range tt.explicitFlags {
				switch name {
				case "timeout":
					args = append(args, "--timeout", tt.timeout.String())
				case "resize-timeout":
					args = append(args, "--resize-timeout", tt.resizeTimeout.String())
				case "modify-timeout":
					args = append(args, "--modify-timeout", tt.modifyTimeout.String())
				}
			}
			if err := fs.Parse(args); err != nil {
				t.Fatalf("failed to parse flags: %v", err)
			}

			gotResize, gotModify := resolveOperationTimeouts(fs, tt.timeout, tt.resizeTimeout, tt.modifyTimeout)
			if gotResize != tt.wantResize {
				t.Errorf("resize timeout = %v, want %v", gotResize, tt.wantResize)
			}
			if gotModify != tt.wantModify {
				t.Errorf("modify timeout = %v, want %v", gotModify, tt.wantModify)
			}
		})
	}
}
