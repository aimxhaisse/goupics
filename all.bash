#!/usr/bin/env bash
#
# author: s. rannou <mxs@sbrk.org>

. config.bash

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
