# Igor

This is a Flowdock bot. It uses AWS DynamoDB and AWS Lambda to check 
if there are new mentions and it responds to them if the date is between some
values stored in DynamoDB configuration.

## No spam

It remembers when was the last communication sent to a particular flowdock nick
to know _not_ to respond to avoid spamming.