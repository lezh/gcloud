// Copyright 2015 Google Inc. All Rights Reserved.
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

// An integration test that uses the real GCS. Run it with appropriate flags as
// follows:
//
//     go test -v -tags integration . -bucket <bucket name>
//
// The bucket's contents are not preserved.
//
// The first time you run the test, it may die with a URL to visit to obtain an
// authorization code after authorizing the test to access your bucket. Run it
// again with the "-oauthutil.auth_code" flag afterward.

// Restrict this (slow) test to builds that specify the tag 'integration'.
// +build integration

package gcs_test

import (
	"flag"
	"log"
	"net/http"
	"testing"

	"github.com/googlecloudplatform/gcsfuse/timeutil"
	"github.com/jacobsa/gcloud/gcs"
	"github.com/jacobsa/gcloud/gcs/gcstesting"
	"github.com/jacobsa/gcloud/gcs/gcsutil"
	"github.com/jacobsa/gcloud/oauthutil"
	"github.com/jacobsa/ogletest"
	"golang.org/x/net/context"
	storagev1 "google.golang.org/api/storage/v1"
)

////////////////////////////////////////////////////////////////////////
// Wiring code
////////////////////////////////////////////////////////////////////////

var fKeyFile = flag.String("key_file", "", "Path to a JSON key for a service account created on the Google Developers Console.")
var fBucket = flag.String("bucket", "", "Empty bucket to use for storage.")

func getHTTPClientOrDie() *http.Client {
	if *fKeyFile == "" {
		panic("You must set --key_file.")
	}

	const scope = storagev1.DevstorageFull_controlScope
	httpClient, err := oauthutil.NewJWTHttpClient(*fKeyFile, []string{scope})
	if err != nil {
		panic("oauthutil.NewJWTHttpClient: " + err.Error())
	}

	return httpClient
}

func getBucketNameOrDie() string {
	s := *fBucket
	if s == "" {
		log.Fatalln("You must set --bucket.")
	}

	return s
}

// Return a bucket based on the contents of command-line flags, exiting the
// process if misconfigured.
func getBucketOrDie() gcs.Bucket {
	// Set up a GCS connection.
	cfg := &gcs.ConnConfig{
		HTTPClient: getHTTPClientOrDie(),
	}

	conn, err := gcs.NewConn(cfg)
	if err != nil {
		log.Fatalf("gcs.NewConn: %v", err)
	}

	// Open the bucket.
	return conn.GetBucket(getBucketNameOrDie())
}

////////////////////////////////////////////////////////////////////////
// Registration
////////////////////////////////////////////////////////////////////////

func TestOgletest(t *testing.T) { ogletest.RunTests(t) }

func init() {
	gcstesting.RegisterBucketTests(func() (deps gcstesting.BucketTestDeps) {
		deps.Bucket = getBucketOrDie()
		deps.Clock = timeutil.RealClock()

		err := gcsutil.DeleteAllObjects(context.Background(), deps.Bucket)
		if err != nil {
			panic("DeleteAllObjects: " + err.Error())
		}

		return
	})
}
