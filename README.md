BEAN
====

A template to build websites in go.

How to
------

Simply fork this project, and build your website upon it.

What it contains
----------------

* a config file for basic settings
* a default template with bootstrap
* a daemonizer to run your website in background (supports chroot)
* examples of configuration files for your web server

Quick start
-----------

    git clone github.com:aimxhaisse/bean
    cd bean
    ./all.bash build
    ./all.bash start

    # browse http://localhost:8080

Real start
----------

This could be automated but you plan to extend this, so I don't want
to add any sort of magic there.

    export project_name="myproj"

    # clone the project
    git clone https://github.com/aimxhaisse/bean.git $project_name
    cd $project_name

    # rename sources/configs to your project's name
    for file in $(find . -name 'bean*'); do git mv $file $(echo $file | sed s/bean/${project_name}/); done

    # edit the variables at the top of all.bash
    $EDITOR all.bash

    # optionally edit the configuration of your website
    $EDITOR root/$project_name.json

    ./all.bash build
    ./all.bash start
    ./all.bash status

    # if this is ok, commit and start to play
    git commit -am "Having setup $project_name from https://github.com/aimxhaisse/bean.git"

Structure
---------

* all.bash is a shell script to manage your website (build, start, stop, ...)
* daemonizer/ contains sources for the daemonizer
* root/ contains dynamic files such as the pid
* root/www/static contains files that are served statically (css, js, images)
* root/www/static contains template files for your pages

Security
--------

It is possible to run your website chrooted (the process won't be able
to leave its directory). For this you need to set CHROOT to "yes" in
all.bash. Also, your code must not rely on absolute paths, use paths
relative to the deployment directory.
