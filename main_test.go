// Copyright 2023 Dolthub, Inc.
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

package main

import (
	"bufio"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_GenerateCalendar(t *testing.T) {
	initializeLocalClient()

	response, err := LambdaHandler(context.Background())
	require.NoError(t, err)

	filepath := "/tmp/oncall.ics"
	writeToFile(t, filepath, response)
}

func initializeLocalClient() {
	// Load the Shared AWS Configuration (~/.aws/config)
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("default"))
	if err != nil {
		panic(err)
	}

	assumedRoleArn, ok := os.LookupEnv("ONCALL_CALENDAR_GENERATOR_ASSUMED_ROLE")
	if !ok {
		panic("ONCALL_CALENDAR_GENERATOR_ASSUMED_ROLE environment variable not set with the role to assume for this test")
	}

	cfg.Region = "us-west-2"
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, assumedRoleArn)

	_, err = provider.Retrieve(ctx)
	if err != nil {
		panic(err)
	}

	assumedRoleConfig := aws.Config{
		Credentials: provider,
		Region:      "us-west-2",
	}

	// Create an SSM Incidents client
	client = ssmcontacts.NewFromConfig(assumedRoleConfig)
}

func writeToFile(t *testing.T, filepath, content string) {
	f, err := os.Create(filepath)
	require.NoError(t, err)
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, w.Flush())
}
