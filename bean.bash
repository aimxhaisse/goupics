#!/usr/bin/env bash
#
# author: s. rannou <mxs@sbrk.org>
#
# A script to manage bean's components.

source config.bash

cd $(dirname $0)

function get-uid {
    return $(id -u)
}

function usage {
    echo -e "usage: $0 [stop|start|restart|status]"
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
	ok "Stopping $PROJECT"
	if [ -f $CHDIR/$PID ]
	then
	    kill -9 $(cat $CHDIR/$PID)
	    rm -f $CHDIR/$PID
	fi
	ok "Daemon stopped"
	exit 0
	;;

    "start")
	if [ $CHROOT = "yes" ]
	then
	    ok "Starting $PROJECT with chroot"
	    ./daemonize -p $PID -j -u $USER -c $CHDIR -- ./$PROJECT -c $CONFIG -l $LOG
	    warn-upon-failure $? "Can't start the daemon, check your config" || exit 1
	else
	    ok "Starting $PROJECT"
	    ./daemonize -p $PID -u $USER -c $CHDIR -- ./$PROJECT -c $CONFIG -l $LOG
	    warn-upon-failure $? "Can't start the daemon, check your config" || exit 1
	fi
	ok "Daemon started"
	exit 0
	;;

    "restart")
	$0 stop
	$0 start
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
esac

usage
exit 1
