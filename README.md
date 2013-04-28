BEAN
===

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
    ./make.bash
    $EDITOR bean.json
    ./bean.rc start

Structure
---------

* bean.bash is a shell script to manage your website
* daemonizer/ contains sources for the daemonizer
* root/ contains dynamic files such as the pid
* root/www/static contains files that are served statically (css, js, images)
* root/www/static contains template files for your pages

Security
--------

It is possible to run your website chrooted (the process won't be able
to leave its directory). For this you need to build your website in
static. Also, your code must not rely on absolute paths, use paths
relative to the deployment directory.
