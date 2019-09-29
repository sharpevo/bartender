#!/bin/bash
BASE="/opt/automation/transmitter"
source $BASE/password.rc
INPUT="$BASE/input"
cd $BASE
./transmitter \
    -sourcepath="$INPUT" \
    -watchpath="$INPUT" \
    -transfer=true \
    -remotepath=/public/home/link/ecs \
    -hostkey="$LOCAL_KEY" \
    -username=xuexh \
    -watch=true \
    -loglevel=debug \
    -password=$LOCAL_PWD
