# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
--- FAIL: TestOfflineBundleReportsFetchFailure (0.00s)
    regression_test.go:32: expected incomplete package, got {ID:regression|manifest JobID:regression IndexPath:/tmp/TestOfflineBundleReportsFetchFailure3692728339/002/index.html PageCount:1 AssetCount:0 FailedPages:[] FailedAssets:[] Incomplete:false ExternalLinks:[]}
FAIL
FAIL	offlinebundle	0.029s
ok  	offlinebundle/cmd/bundle	0.001s
ok  	offlinebundle/internal/builder	0.002s
ok  	offlinebundle/internal/domain	0.001s
ok  	offlinebundle/internal/fetcher	0.005s
ok  	offlinebundle/internal/report	0.002s
ok  	offlinebundle/internal/service	0.011s
ok  	offlinebundle/internal/storage	0.009s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/bundle): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/bundle): exit `0`
