build-AwsIncidentManagerOncallCalendar:
	GOOS=linux GOARCH=arm64 go build -o bootstrap
	zip dolthub-prod-ssm-contact-ics-function.zip bootstrap