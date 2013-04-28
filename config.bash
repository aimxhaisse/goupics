#!/usr/bin/env bash
#
# author: s. rannou <mxs@sbrk.org>

# edit these to your needs
PROJECT="bean"		# name of the project
CHROOT="no"		# [yes|no]
USER="mxs"		# used for daemonization

# edit these to your needs if you know what you are doing
CHDIR="root"		# directory where we chdir/chroot
PID="$PROJECT.pid"	# relative to $CHDIR
BIN="$PROJECT"		# relative to $CHDIR
LOG="$PROJECT.log"	# relative to $CHDIR
CONFIG="$PROJECT.json"	# relative to $CHDIR

# Few tools used by all script
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
