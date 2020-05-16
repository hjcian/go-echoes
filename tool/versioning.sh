#!/bin/bash
scriptdir=`dirname "$(realpath $0)"`
projectdir=`dirname ${scriptdir}`

ver=`cat ${projectdir}/VERSION`

git tag -a $ver -m "tagging version"
RESULT=$?

if [ $RESULT == 0 ]; then
    git push origin $ver
else
    echo "you should upgrade the version then do again."
fi