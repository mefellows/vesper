# Serverless example

```
GOOS=linux go build -o bin/main
sls invoke local -f hello --data '{"Username": "me", "Password":"securething"}'
```
