# Releasing

Releases are performed by pushing to Github with an appropriate tag.

As we use semantic commits and versions, we can easily generate a useful changelog.

```
make release
# If successful, the command will instruct you to push to Github
git push --follow-tags
```

That's it! :)