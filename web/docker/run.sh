#!/usr/bin/env bash

docker run \
    --name ebid-nginx \
    -v /Users/cirelli/Projects/Steve/ebidlocal.com_playground/html:/usr/share/nginx/html:ro \
    -v /Users/cirelli/Projects/Steve/ebidlocal.com_playground/configs/nginx.conf:/etc/nginx/nginx.conf:ro \
    -v /Users/cirelli/Projects/Steve/ebidlocal.com_playground/configs/sites-available:/etc/nginx/sites-enabled:ro \
    -v /Users/cirelli/Projects/Steve/ebidlocal.com_playground/certs/private:/etc/ssl/private:ro \
    -v /Users/cirelli/Projects/Steve/ebidlocal.com_playground/certs/public:/etc/ssl/certs:ro \
    -p 8383:80 \
    -p 8443:443 \
    --rm \
    nginx
