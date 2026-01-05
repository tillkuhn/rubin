#!/usr/bin/env bash
# test handler for polly callback demo
sleepy=0 # simulate long processing

me=$(basename "$0")
# or if you only want to read from stdin into a variable
if [ ! -t 0 ]; then
    echo "$me: Reading data from STDIN"  # cat cloud-event.json | ./script.sh
    json=$(</dev/stdin)
    echo "$me: Received message: $json, processing for $sleepy seconds"
    sleep $sleepy # test shutdown handling error="signal: killed"
    echo "$me: Processing finished after $sleepy seconds"
       if ! jq type --argjson data "$json" >/dev/null; then
        echo "$me: Message is not valid JSON, stop further processing"; exit 0
    fi
    event_type=$(jq -n --argjson data "$json" '$data.type')
    subject=$(jq -n --argjson data "$json" '$data.subject')

    echo "$me: Processing $event_type event for subject $subject"
else
    echo "$me: Running interactively"
fi

# https://stackoverflow.com/a/7045517/4292075
#while read line; do
#  echo "... $line"
#done < "${1:-/dev/stdin}"
