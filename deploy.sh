#!/bin/sh
echo '================ Build process starting! ================'
echo "-> Installing all dependencies\n"

cat dependencies.txt | awk '{print "echo Installing "$1"... & go get -u "$1"\0"}' |  xargs -0 bash -c

while [ "$1" != "" ]; do
    PARAM=`echo $1 | awk -F= '{print $1}'`
    VALUE=`echo $1 | awk -F= '{print $2}'`
    case $PARAM in
        --build)
            if [ "$VALUE" = "" ]
            then
                VALUE=bin/rabbitmq-monitor
            fi
            echo "-> Building CLI Tool to $VALUE..."
            go build -o $VALUE
            echo "== Deploy & Build done successfully. bin stored on $VALUE"
            exit 0
            ;;
        *)
            echo "ERROR: unknown parameter \"$PARAM\""
            exit 1
            ;;
    esac
    shift
done
echo '===== Deploy process successfully done without build ====='
exit 0