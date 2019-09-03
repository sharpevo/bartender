#!/bin/bash
source ~/go/src/automation/password.rc
INPUT="input"
inotifywait -m -r -q $INPUT -e close_write --format '%w%f'|while read newfile
do
    echo "==> $newfile"
../transmitter \
    -sourcepath="$newfile" \
    -watchpath="$INPUT" \
    -transfer=true \
    -remotepath=/root/wuy/testauto \
    -hostkey="$ECS_KEY" \
    -username=root \
    -watch=true \
    -loglevel=debug \
    -loglevel=debug \
    -password=$ECS_PWD
done
