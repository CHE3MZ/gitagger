# FAQ & troubleshooting

## It says "already on vX"

HEAD already has a tag. That's the double-tag guard — make a new commit, or retag deliberately with `-f`.

## It says the tag "already exists on remote"

Someone (or a previous run) already pushed that tag. `gitagger doctor` shows local-only vs remote-only tags. Overwrite only if you mean it (`-f` force-pushes).

## It refuses to move a float backwards

`release-action.sh` only moves `vX` / `vX.Y` forward. If it refuses, the new tag doesn't descend from the floats — usually a rewound or stale history. Don't force it blindly: verify the new tip contains everything (`git diff <float> <tip> --stat`), then move floats by hand with `git tag -f` + `git push --force`. And stop rewriting public history — merge forward instead.

## It kept the tag locally

No remote (`no remote "origin"`) or unreachable remote (offline?) — the tag is created, just not pushed. Run the printed `git push <remote> <tag>` later.

## It tagged nothing on a release commit

Correct behavior, not breakage: release commits already carry tags, so gitagger aborts instead of double-tagging. The action's self-test hits this on purpose and probes a fresh commit instead.

## My format wasn't detected

Unknown tag styles are ignored; with no recognizable tags gitagger starts at triple. Force a style with `--format`, or set `format:` in config. `doctor` lists "odd tags" it skipped.

## Config ignored?

`gitagger check -v` shows the resolved values and the file they came from. Remember: flags beat config, and the file is looked up in the *working directory* (`-p` / `working-directory` / `-w` change where that is).

## Typo'd command?

`hlp` and friends get a did-you-mean suggestion with exit code 3. Gibberish gets none.

## Still stuck?

Open an issue with `NO_COLOR=1 gitagger check -v` output, your `.gitagger.yml` (secrets redacted), and the exact command plus its output.
