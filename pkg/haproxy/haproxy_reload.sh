#!/bin/bash
set -e

reload_haproxy(){
    if [ ! -f /var/run/haproxy.pid ]; then
        if service haproxy start; then
            return 0
        else
            return 1
        fi
    fi
    # restart service
    if haproxy -p /var/run/haproxy.pid -f $1 -sf $(cat /var/run/haproxy.pid); then
        return 0
    else
        return 1
    fi
}

apply_config()
{
    # apply new config
    if [ $2 == "start" ]; then
        echo "starting haproxy"
        reload_haproxy $1
    elif ! cmp -s $1 /etc/haproxy/haproxy_new.cfg  ; then
        echo "reloading haproxy config with the new config changes"
        # replace old config
        cp -r /etc/haproxy/haproxy_new.cfg  $1
        reload_haproxy $1
    else
        echo "no changes in haproxy config"
        return 0
    fi
}

apply_config $1 $2