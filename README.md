AWS Incident Manager On-call Calendar Generator
===

AWS Lambda function that generates an on-call schedule that calendar apps can subscribe to. 

To run in AWS Lambda, make sure: 
- your Lambda execution role has permission to call [`ssm-contacts:ListRotationShifts`](https://docs.aws.amazon.com/incident-manager/latest/APIReference/API_SSMContacts_ListRotationShifts.html)
- you populate the `ROTATION_ID_ARN` environment variable in your Lambda function's environment with the ARN of the AWS Incident Manager rotation for which you want to generate a calendar.

To test locally, see the `Test_GenerateCalendar` function in `main_test.go` and make sure:
- you populate the `ONCALL_CALENDAR_GENERATOR_ASSUMED_ROLE` environment variable with the role you need to assume in order to access. For DoltHub developers, this should be the ARN of the LiquidataDeveloper role.
- you may need to alter the `initializeLocalClient` code in order to load the correct AWS credentials for your environment.