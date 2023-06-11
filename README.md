# CloudTrail Query

Tool to run SQL queries to AWS CloudTrail and print results as json objects. 

The tool leverages the [AWS CloudTrail Lake](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-lake.html) functionality, which allows you to run SQL queries against CloudTrail Events. Be sure to have at least an [event datastore created in CloudTrail](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/query-event-data-store.html) to allow querying

The tool outputs a series of json objects to stdout, each json representing a resulting row

## Running

This software uses docker, so its necessary to have a docker engine installed

Use the [`query.sh`](query.sh) helper script to automatically pull the latest image and run queries. No need to clone or build this repository

## Development

Use `make` or `make build` to build the docker image + `./query.sh` to issue queries

You can run locally directly with `go run ./cmd` as well

## Examples

Fetch some information about the 5 last AwsApiCall events

```
% ./query.sh $'select eventTime, eventName, eventSource from ad85f16b-ed48-4cd0-8833-e0d5e6b22725 where eventTime > \'2023-06-10 00:00:00.000\' and eventType = \'AwsApiCall\' order by eventTime desc limit 10' | jq .

2023/06/11 16:27:02 running query: select eventTime, eventName, eventSource from ad85f16b-ed48-4cd0-8833-e0d5e6b22725 where eventTime > '2023-06-10 00:00:00.000' and eventType = 'AwsApiCall' order by eventTime desc limit 5
2023/06/11 16:27:04 query id: 75d844cb-192a-4280-8b22-b442aa49bdc7
2023/06/11 16:27:22 progress: 5/5 (100.00%)
{
  "eventName": "DescribeRepositories",
  "eventSource": "ecr.amazonaws.com",
  "eventTime": "2023-06-11 16:26:39.000"
}
{
  "eventName": "AssumeRole",
  "eventSource": "sts.amazonaws.com",
  "eventTime": "2023-06-11 16:26:37.000"
}
{
  "eventName": "DescribeInstanceHealth",
  "eventSource": "elasticloadbalancing.amazonaws.com",
  "eventTime": "2023-06-11 16:26:29.000"
}
{
  "eventName": "DescribeInstances",
  "eventSource": "ec2.amazonaws.com",
  "eventTime": "2023-06-11 16:26:28.000"
}
{
  "eventName": "GetServiceQuota",
  "eventSource": "servicequotas.amazonaws.com",
  "eventTime": "2023-06-11 16:26:27.000"
}
```

## Caveat: CloudTrail Events Format

CloudTrail events can be problematic because they can contain information that _looks_ like json but it is not actually json. This tool tries to compensate for that by parsing json-like events into actual json objects

Example:

A raw userIdentity column persisted on CloudTrail:
```
"userIdentity: "{type=AssumedRole, principalid=XXXXXXXXXXXXXXXXXXXXX:secrets-provider, arn=arn:aws:sts::XXXXXXXXXXXX:assumed-role/some-role/secrets-provider, accountid=XXXXXXXXXXXX, accesskeyid=XXXXXXXXXXXXXXXXXXXX, username=null, sessioncontext={attributes={creationdate=2023-06-07 23:59:12.000, mfaauthenticated=true}, sessionissuer={type=Role, principalid=XXXXXXXXXXXXXXXXXXXXX, arn=arn:aws:iam::XXXXXXXXXXXX:role/some-role, accountid=XXXXXXXXXXXX, username=some-role}, sourceidentity=null, ec2roledelivery=null, ec2issuedinvpc=null}, invokedby=null, identityprovider=null, credentialid=null, onbehalfof=null}"
```

The resulting json from the tool (such values have the "__parsed" suffix):
```
"userIdentity__parsed": {
    "accesskeyid": "XXXXXXXXXXXXXXXXXXXX",
    "accountid": "XXXXXXXXXXXX",
    "arn": "arn:aws:sts::XXXXXXXXXXXX:assumed-role/some-role/secrets-provider",
    "credentialid": null,
    "identityprovider": null,
    "invokedby": null,
    "onbehalfof": null,
    "principalid": "XXXXXXXXXXXXXXXXXXXXX:secrets-provider",
    "sessioncontext": {
      "attributes": {
        "creationdate": "2023-06-07 23:59:12.000",
        "mfaauthenticated": true
      },
      "ec2issuedinvpc": null,
      "ec2roledelivery": null,
      "sessionissuer": {
        "accountid": "XXXXXXXXXXXX",
        "arn": "arn:aws:iam::XXXXXXXXXXXX:role/some-role",
        "principalid": "XXXXXXXXXXXXXXXXXXXXX",
        "type": "Role",
        "username": "some-role"
      },
      "sourceidentity": null
    },
    "type": "AssumedRole",
    "username": null
  },
```

Use the arg `-E` to disable this behavior and only print raw events

