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
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"

	ics "github.com/arran4/golang-ical"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
)

const weekDuration = time.Hour * 24 * 7

var client *ssmcontacts.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client = ssmcontacts.NewFromConfig(cfg)
}

func main() {
	lambda.Start(LambdaHandler)
}

func LambdaHandler(ctx context.Context) (string, error) {
	shifts, err := loadOncallShifts(ctx)
	if err != nil {
		return "", err
	}

	cal := createOnCallCalendar(shifts)
	// TODO: Some clients may need us to set content-type in order to recognize the ICS file, but so far in testing,
	//       Mac's Calendar app works just fine without it.
	//       https://stackoverflow.com/questions/46171369/how-to-change-aws-lambda-content-type-from-plain-text-to-html
	return cal.Serialize(), nil
}

func loadOncallShifts(ctx context.Context) ([]types.RotationShift, error) {
	rotationIdArn, ok := os.LookupEnv("ROTATION_ID_ARN")
	if !ok {
		return nil, fmt.Errorf("ROTATION_ID_ARN environment variable not set with the on-call rotation to query")
	}

	// Get the first page of results for the oncall schedule
	output, err := client.ListRotationShifts(ctx, &ssmcontacts.ListRotationShiftsInput{
		EndTime:    aws.Time(time.Now().Add(12 * weekDuration)),
		RotationId: aws.String(rotationIdArn),
		StartTime:  aws.Time(time.Now().Add(-1 * weekDuration)),
	})
	if err != nil {
		return nil, err
	}

	return output.RotationShifts, nil
}

func createOnCallCalendar(shifts []types.RotationShift) *ics.Calendar {
	cal := ics.NewCalendar()
	// MethodPublish should be used for ICS files that contain the full set of calendar events
	// https://stackoverflow.com/questions/28552946/ics-ical-publish-request-cancel
	cal.SetMethod(ics.MethodPublish)
	for _, shift := range shifts {
		// use the shift start time as the persistent, unique identifier for this event
		event := cal.AddEvent(fmt.Sprintf("%d", shift.StartTime.UTC().Unix()))

		// TODO: Add information about shift overrides. shift.Type specifies if this is a regular
		//       shift or an overridden shift, and shift.ShiftDetails.OverriddenContactIds shows
		//       the overridden contact.

		// Create an all day event, instead of using the exact start/end time, so that
		// the event displays cleaner on calendars
		event.SetAllDayStartAt(*shift.StartTime)
		event.SetAllDayEndAt(*shift.EndTime)

		splits := strings.Split(shift.ContactIds[0], "/")
		// TODO: strings.Title is deprecated; should move to golang.org/x/text/cases
		oncall := strings.Title(splits[len(splits)-1])

		event.SetSummary("On-Call: " + oncall)
		event.SetDescription(fmt.Sprintf("%s is on-call for DoltHub", oncall))
	}

	return cal
}
