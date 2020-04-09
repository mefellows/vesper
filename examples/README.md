# Serverless example

```
GOOS=linux go build -o bin/main
sls invoke local -f middleware --data '{"Username": "me", "Password":"securething"}'
```
