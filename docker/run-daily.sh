#!/bin/sh
set -eu

run_with_env() {
	label=$1
	file=$2
	if [ ! -r "$file" ]; then
		echo "neurofeed-cron: missing or unreadable env file: $file (profile=$label)" >&2
		exit 1
	fi
	echo "neurofeed-cron: start profile=$label file=$file"
	set -a
	# shellcheck disable=SC1090
	. "$file"
	set +a
	/usr/local/bin/neurofeed
	echo "neurofeed-cron: done profile=$label"
}

case "${1:-}" in
default) run_with_env default /config/.env ;;
vanessa) run_with_env vanessa /config/.env.vanessa ;;
*)
	run_with_env default /config/.env
	run_with_env vanessa /config/.env.vanessa
	;;
esac
