## Git setup

Add the following to `.git/hooks/pre-commit` and make executable

```
#!/bin/bash
ROOT=`git rev-parse --show-toplevel`
$ROOT/set_build_no.py
git add fvserver/version.plist
```
