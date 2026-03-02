#!/usr/bin/env python3
import sys, json, subprocess, re
from datetime import time
from zoneinfo import ZoneInfo
from datetime import datetime

HELP = """This hook prevents committing during working hours unless the commit message contains
an exception tag referencing a Balance Log entry (e.g., 'EXCEPTION: EX-YYYYMMDD-01').
Set VLK_ALLOW_DAYTIME_COMMIT=1 to bypass (remember to log/offset)."""

def now_in_window(dt, workdays, start_hour, end_hour):
    if dt.weekday() not in workdays:
        return False
    start = dt.replace(hour=start_hour, minute=0, second=0, microsecond=0)
    end = dt.replace(hour=end_hour, minute=0, second=0, microsecond=0)
    return start <= dt < end

def main():
    if len(sys.argv) < 3:
        print("Usage: validate_hours.py commit-msg <msgfile> <config_json>", file=sys.stderr)
        sys.exit(2)
    mode = sys.argv[1]
    msgfile = sys.argv[2]
    cfg_path = sys.argv[3] if len(sys.argv) >= 4 else None
    if mode != "commit-msg":
        # Only implement commit-msg for simplicity
        sys.exit(0)

    with open(cfg_path, "r") as f:
        cfg = json.load(f)
    tz = ZoneInfo(cfg.get("timezone","America/New_York"))
    workdays = cfg.get("workdays",[0,1,2,3,4])
    start_hour = int(cfg.get("start_hour",9))
    end_hour = int(cfg.get("end_hour",17))
    tag_prefix = cfg.get("exception_tag_prefix","EXCEPTION:")
    allow_env = cfg.get("allow_env_var","VLK_ALLOW_DAYTIME_COMMIT")

    # Get author time of the commit that is being created (HEAD isn't updated yet in commit-msg)
    # Use local 'now' as an approximation since author date will default to 'now'.
    now_local = datetime.now(tz)

    # Read commit message
    with open(msgfile, "r", encoding="utf-8") as f:
        msg = f.read()

    has_exception = tag_prefix in msg

    import os
    if os.environ.get(allow_env,"") == "1":
        sys.exit(0)

    if now_in_window(now_local, workdays, start_hour, end_hour) and not has_exception:
        print("✖ Commit blocked during working hours ({}).".format(now_local.strftime("%Y-%m-%d %H:%M %Z")), file=sys.stderr)
        print("   Add an exception tag (e.g., '{} EX-YYYYMMDD-01') referencing your Balance Log entry,".format(tag_prefix), file=sys.stderr)
        print("   or set {}=1 to override (remember to log an offset).".format(allow_env), file=sys.stderr)
        print("\n" + HELP, file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
