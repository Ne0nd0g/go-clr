# Continuous integration

GitHub Actions builds the library and all examples on Windows for amd64 and 386,
runs vet and tests, and scans the full module with govulncheck. CodeQL uses a
manual Windows amd64 build so it includes the Windows-tagged sources. Both
workflows run on pushes and pull requests to `dev` and `main`, manual dispatch,
and weekly schedules. Workflow edits also trigger CI.

Go comes from `go.mod`; checkout, setup-go, CodeQL, and golangci-lint action
versions follow the Merlin workspace template. The lint binary is pinned to
v2.14.0 so baseline results remain reproducible.

## Existing findings

Lint (including gosec) checks findings on added or changed lines since the fixed pre-CI commit
`02d72d3519181e1dc71b409bf8eced508cc44537`. The baseline does not advance with
each PR, so new findings remain gated after merge. Full checkout history is
required. Existing findings are deferred, not resolved: these include unchecked
errors, overwritten errors, deprecated syscalls, unused COM vtable slots, and
unsafe-pointer diagnostics. Do not remove vtable slots simply to satisfy lint;
their positions define the COM ABI.

Standalone vet disables `unsafeptr` because existing return-value conversions
fail it; golangci-lint still checks that analyzer on changed lines. Tests disable
their implicit vet invocation because vet runs separately. Gosec excludes G103
(unsafe operations intrinsic to the bindings) and G115 (the workspace's integer
conversion exclusion); other rules remain enabled.

Run the complete lint inventory on Windows, without the CI baseline filter:

```powershell
golangci-lint run --max-same-issues=0 --max-issues-per-linter=0 ./...
```

On a non-Windows development host, set `GOOS=windows` and `GOARCH=amd64` before
building or analyzing. Windows executables and the complete test suite require a
Windows host.

## Validation limits

Tests cover UTF-16 encoding/decoding, entry-point parameter detection, HRESULT
diagnostics, and Windows SAFEARRAY creation, bounds, byte copying, element writes,
out-of-bounds errors, and access/unaccess cleanup. CI runs the tests uncached with
coverage reporting on both architectures. The SAFEARRAY tests call Windows APIs
using inert byte buffers and require no external assemblies or .NET runtime.

The first test run at `25fdcf5` passed on amd64 with 19.8% package coverage but
found a missing `ntdll.dll!RtlCopyMemory` export on 386. `CreateSafeArray` now uses
[RtlMoveMemory](https://learn.microsoft.com/en-us/windows/win32/devnotes/rtlmovememory)
with the same source, destination, and byte count. It returns no error status,
so the copy does not interpret GetLastError as a failure. The tests also check
that later changes to the source bytes do not alter the SAFEARRAY contents.

These tests do not demonstrate managed assembly execution or correctness of every
COM interface. CodeQL extraction has passed on GitHub Actions; native execution
of the memory-copy fix needs verification after the maintainer commits and pushes
it. This library has no release workflow or binary distribution step; tagging and
publishing remain manual.
