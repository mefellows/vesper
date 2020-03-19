# Warmup middleware

Implements a warmup handler for https://www.npmjs.com/package/serverless-plugin-warmup

Example invocation for `fn`:

```
GOOS=linux go build -o bin/main
sls invoke local -f hello --data '{"Username": "me", "Password":"securething"}'
```
