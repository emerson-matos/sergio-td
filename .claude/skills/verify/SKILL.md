---
name: verify
description: Run all server tests to verify changes are correct
---

Run the Go server test suite to verify that changes haven't broken anything.

```bash
cd server && go test -v ./...
```

If any tests fail, analyze the output and fix the issues. Re-run until all tests pass.
