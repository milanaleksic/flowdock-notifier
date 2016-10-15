# Igor

[![Build Status](https://semaphoreci.com/api/v1/milanaleksic/igor/branches/master/badge.svg)](https://semaphoreci.com/milanaleksic/igor)

This is a Flowdock bot. It uses AWS DynamoDB and AWS Lambda to check 
if there are new mentions and it responds to them if the date is between some
values stored in DynamoDB configuration.

## No spam

It remembers when was the last communication sent to a particular flowdock nick
to know _not_ to respond to avoid spamming.

## How to

### ... deploy

0. Make sure you have:
    - Docker and JQ packages installed
    - AWS account
    - Flowdock account
    - Client ID of a Google+ API approved Google Dev Console application
1. rename `personal.env.template` to `personal.env` and introduce adequate values (ignore the ones with `GENERATED` prefix)
2. run `cmd/main/cf/non_supported_pre.sh` to update the `personal.env` with the generated values

Deploy to AWS with `make form`. Update function only with `make update`. Invoke directly via `make invoke`.

### ... configure

For now, until phase 3 is done, you need to manually enter following values in the `igor-config` table:

- id=`message`, value=`Hi, I am unavailable from {{.From}} until {{.Until}}. It might be I don't answer your message until then.`
- id=`activeFrom`, value=`14 Oct 16 06:00 UTC`
- id=`activeUntil`, value=`18 Oct 16 06:00 UTC`
