# Warmup middleware

Implements a warmup handler for https://www.npmjs.com/package/serverless-plugin-warmup

Example invocation for `fn`:

```
sls invoke -f fn -d '{ "Event": { "source": "serverless-plugin-warmup" } }'
```
