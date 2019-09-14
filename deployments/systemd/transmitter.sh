#!/bin/bash
BASE="/opt/automation/transmitter"
source $BASE/password.rc
INPUT="$BASE/input"
cd $BASE
inotifywait -m -r -q $INPUT -e close_write --format '%w%f'|while read newfile
do
    echo "==> $newfile"
./transmitter \
    -sourcepath="$newfile" \
    -watchpath="$INPUT" \
    -transfer=true \
    -remotepath=/public/home/link/ecs \
    -hostkey="$LOCAL_KEY" \
    -username=xuexh \
    -watch=false \
    -password=$LOCAL_PWD
#-loglevel=debug \
#-namepattern=".\.abc$" \
done
