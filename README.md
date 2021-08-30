# slarf
Slarf is a simple tool that accepts Slack client user credentials (the `d` cookie paired with an `xoxc-` token) to dump the user directory for the workspace tied to the token.

## Usage
```shell
$ ./slarf -cookie aX9QnD8F[REDACT] -token xoxc-1300035[REDACT] | jq '.[].name'
"slackbot"
"actae0n"
[ SNIP ]
```

## TODO
Draw the rest of the owl. I'm gonna add support for dumping full channels, attachments, etc.
