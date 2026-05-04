#!/bin/sh
set -eu
# Cron runs jobs with a minimal environment; snapshot full env at start (Railway / Compose)
# so the 07:00 job sees the same variables as PID 1.
/usr/bin/bash -c 'export -p > /run/neurofeed-cron-env.sh'
chmod 600 /run/neurofeed-cron-env.sh
exec "$@"
