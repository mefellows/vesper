# Changelog
Do this to generate your change history

    git log --pretty=format:'  * [%h](https://github.com/mefellows/vesper/commit/%h) - %s (%an, %ad)' vX.Y.Z..HEAD | egrep -v "wip(:|\()" | grep -v "docs(" | grep -v "chore(" | grep -v Merge | grep -v "test("

## Versions

