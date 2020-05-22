# Changelog
Do this to generate your change history

    git log --pretty=format:'  * [%h](https://github.com/mefellows/vesper/commit/%h) - %s (%an, %ad)' vX.Y.Z..HEAD | egrep -v "wip(:|\()" | grep -v "docs(" | grep -v "chore(" | grep -v Merge | grep -v "test("

## Versions

### v0.0.0 (22 May 2020)
  * [d4ff8bc](https://github.com/pact-foundation/pact-go/commit/d4ff8bc) - docs: add coverage badge on master (Matt Fellows, Fri Apr 17 12:40:44 2020 +1000)
  * [7dd3157](https://github.com/pact-foundation/pact-go/commit/7dd3157) - docs: format go examples (Matt Fellows, Fri Apr 17 12:39:20 2020 +1000)
  * [1c4cfb2](https://github.com/pact-foundation/pact-go/commit/1c4cfb2) - feat: parser middlewares and ability to turn off auto unmarshalling (#4) (bengglim, Thu Apr 9 17:23:58 2020 +1000)
  * [ba7383a](https://github.com/pact-foundation/pact-go/commit/ba7383a) - chore: update example formatting in docs (Matt Fellows, Fri Apr 3 10:10:29 2020 +1100)
  * [4c32646](https://github.com/pact-foundation/pact-go/commit/4c32646) - chore: update example formatting in docs (Matt Fellows, Fri Apr 3 10:08:48 2020 +1100)
  * [999ad34](https://github.com/pact-foundation/pact-go/commit/999ad34) - chore: add screencast (Matt Fellows, Fri Apr 3 10:07:21 2020 +1100)
  * [1b2a23d](https://github.com/pact-foundation/pact-go/commit/1b2a23d) - chore: update docs (Matt Fellows, Fri Apr 3 10:06:20 2020 +1100)
  * [6677edc](https://github.com/pact-foundation/pact-go/commit/6677edc) - docs: add asciicast and better docs (Matt Fellows, Fri Apr 3 10:03:14 2020 +1100)
  * [3d3130a](https://github.com/pact-foundation/pact-go/commit/3d3130a) - fix: set correct logger in serverless template (Matt Fellows, Fri Apr 3 00:23:08 2020 +1100)
  * [7de4173](https://github.com/pact-foundation/pact-go/commit/7de4173) - chore: update docs, makefile and travis build (Matt Fellows, Fri Apr 3 00:22:08 2020 +1100)

