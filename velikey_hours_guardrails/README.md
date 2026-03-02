
# VeliKey Hours Guardrails (Optional)

This package adds a lightweight `commit-msg` hook that **blocks commits during working hours**
(default 9:00–17:00 America/New_York, Mon–Fri) unless your commit message includes an
exception tag referencing a Balance Log entry, e.g.:

    EXCEPTION: EX-20250115-01

## Install
1) Copy `.githooks` and `scripts` into your repo root (or keep them in a shared dotfiles repo).
2) Point Git to the hooks path (once per repo):
   ```bash
   git config core.hooksPath .githooks
   chmod +x .githooks/commit-msg
   ```
3) Adjust working hours/timezone in `scripts/velikey_hours_config.json`.

## Using exceptions
During Normal Working Hours, either:
- Include `EXCEPTION: EX-YYYYMMDD-XX` in your commit message (and log/offset the time), or
- Temporarily override with `VLK_ALLOW_DAYTIME_COMMIT=1` for one-off commits (still log/offset).

## Notes
- This hook avoids generating conflicting records in the first place.
- Do **not** rewrite timestamps to conceal activity. If a legitimate exception occurs,
  log it and tag the commit message.
- The hook checks local time-of-commit (author date defaults to now). If your workflow
  sets custom author dates, adjust the script accordingly.
