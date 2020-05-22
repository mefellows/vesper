# Releasing

Releases are performed by pushing to Github with an appropriate tag.

As we use semantic commits and versions, we can easily generate a useful changelog.

```
make release
# If successful, the command will instruct you to push to Github
git push --follow-tags
```

The whole process should look something like this:

```
➜  vesper git:(master) ✗ make release
 -----> Releasing Vesper 🚀
        Finding current version
        Increment 'minor' version from v0.0.0 to v0.1.0
 -----> Generating changelog
        Updating CHANGELOG.md
276
1895
        Changelog updated
 -----> Committing changes
        unstaging files
Unstaged changes after reset:
M	CHANGELOG.md
M	scripts/release.sh
        adding CHANGELOG.md
        commiting
[master 65bedf1] chore(release): release v0.1.0
 1 file changed, 12 insertions(+)
 -----> Creating tag v0.1.0
        Done - check your git logs, CHANGELOG, and then run 'git push --follow-tags'.
```

That's it! :)