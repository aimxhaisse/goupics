#!/usr/bin/env bash
#
# author: s. rannou <mxs@sbrk.org>

# edit these to your needs

# name of the project
# don't put / in it because it's used to create files.
PROJECT="goupics"

# enable/disable chroot
# possible values are "yes" or "no".
# if you enable chroot, you need to start this script as root
CHROOT="yes"

# user to run as
USER="mxs"

# edit these to your needs if you know what you are doing
CHDIR="root"			# directory where we chdir/chroot
PID="$PROJECT.pid"		# relative to $CHDIR
BIN="$PROJECT"			# relative to $CHDIR
LOG="$PROJECT.log"		# relative to $CHDIR
CONFIG="$PROJECT.json"		# relative to $CHDIR

cd $(dirname $0)

function ok {			# <msg>
    echo -e "\033[0;32;49m$@\\033[0m"
    return 0
}

function ko {			# <msg>
    echo -e "\033[0;31;49mError: $@\\033[0m"
    return 1
}

function title {		# <msg>
    echo -e "\033[0;32;40m$@\\033[0m"
    return 0
}

function warn-upon-failure {	# <return_value> <msg>
    local ret=$1
    local msg=$2
    if [ $ret -ne 0 ]
    then
	ko "$msg"
    fi
    return $ret
}

function be-quiet {		# <command1>
    $@ 2>&1 > /dev/null
    return $?
}


function get-uid {
    return $(id -u)
}

function usage {
    echo -e "usage: $0 [build|stop|start|restart|status|convert]"
}

function check-env {
    # make sure $DEPLOY_DIR exists
    if ! [ -d $DEPLOY ]
    then
	warn-upon-failure 1 "$DEPLOY does not exist, maybe you should edit \$DEPLOY" || return 1
    fi

    # make sure $USER exists
    be-quiet id $USER
    warn-upon-failure $? "User $USER does not exist, maybe you should edit \$USER"

    return 0
}

check-env || exit 1

case $1 in
    "stop")
	ok "stopping $PROJECT"
	if [ -f $CHDIR/$PID ]
	then
	    kill -9 $(cat $CHDIR/$PID)
	    rm -f $CHDIR/$PID
	fi
	ok "daemon stopped"
	exit 0
	;;

    "start")
	if [ $CHROOT = "yes" ]
	then
	    ok "starting $PROJECT with chroot"
	    ./daemonize -p $PID -j -u $USER -c $CHDIR -- ./$PROJECT -c $CONFIG -l $LOG
	    warn-upon-failure $? "Can't start the daemon, check your config" || exit 1
	else
	    ok "starting $PROJECT"
	    ./daemonize -p $PID -u $USER -c $CHDIR -- ./$PROJECT -c $CONFIG -l $LOG
	    warn-upon-failure $? "Can't start the daemon, check your config" || exit 1
	fi
	ok "daemon started"
	exit 0
	;;

    "restart")
	$0 stop
	$0 start
	exit 0
	;;

    "convert")
	if [ $# -ne 2 ]
	then
	    echo "usage: $0 convert [target-directory]"
	    exit 1
	fi
	target=$2
	for file in $target/*
	do
	    echo $file
	    convert $file -quality 90 -resize 1920x1080 ${file}.resized
	    mv ${file}.resized $file
	done
	exit 0
	;;

    "status")
        pid=$(ps -o pid= --pid $(cat $CHDIR/$PID 2>/dev/null) 2>/dev/null)
        if [ "$pid" != "" ]
        then
            ok "$PROJECT is running, with pid $pid"
        else
            ko "$PROJECT is down"
        fi
        exit 0
        ;;

    "log")
	if [ -f $CHDIR/$LOG ]
	then
	    ok "last 25 log entries:"
	    tail -n 25 $CHDIR/$LOG
	else
	    ko "can't find log file ($CHDIR/$LOG)"
	fi
	exit 0
	;;

    "build")
	# build the daemonizer
	if [ ! -f daemonize ] || [ $(stat -c %Y daemonize) -lt $(stat -c %Y daemonizer/daemonizer.c) ]
	then
	    ok "building daemonize..."
	    gcc -W -Wall -pedantic -ansi -O3 -Wno-unused-result daemonizer/daemonizer.c -o daemonize
	    warn-upon-failure $? "can't build daemonize"
	    ok "daemonize built (or not)"
	else
	    ok "daemonize is already up-to-date (skipped)"
	fi
	
	# build the website
	if [ $CHROOT = "yes" ]
	then
	    ok "building $PROJECT..."
	    CGO_ENABLED=0 go build -o $CHDIR/$PROJECT -a -ldflags -d
	    warn-upon-failure $? "unable to statically build $PROJECT"
	else
	    ok "building $PROJECT..."
	    go build -o $CHDIR/$PROJECT
	    warn-upon-failure $? "unable to build $PROJECT"
	fi
	ok "$PROJECT built (or not)..."
	exit 0
	;;
esac

usage
exit 1
