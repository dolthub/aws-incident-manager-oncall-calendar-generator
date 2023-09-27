build-AwsIncidentManagerOncallCalendar:
	GOOS=linux go build -o bootstrap
	zip dolthub-prod-ssm-contact-ics-function.zip bootstrap