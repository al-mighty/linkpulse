#!/bin/bash
set -e

CONF="/opt/cheslav/nginx/conf.d/cheslav.conf"

# Connect to cheslav-net
docker network connect cheslav-net linkpulse-app-1 2>/dev/null || echo "already connected"

# Add nginx location
if grep -q "linkpulse" "$CONF"; then
    echo "linkpulse already in nginx config"
else
    echo "Adding linkpulse to nginx config..."
    sed -i '$ d' "$CONF"
    cat >> "$CONF" << 'EOF'

    location /linkpulse/ {
        proxy_pass http://linkpulse-app-1:8080/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
EOF
    echo "nginx config updated"
fi

docker exec cheslav-nginx nginx -t && docker restart cheslav-nginx

sleep 2
curl -sk https://localhost/linkpulse/api/health && echo ""
echo "Live: https://cheslav.space/linkpulse/app"