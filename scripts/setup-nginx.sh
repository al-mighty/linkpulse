#!/bin/bash
set -e

CONF="/opt/cheslav/nginx/conf.d/cheslav.conf"

# Connect to cheslav-net
docker network connect cheslav-net linkpulse-app-1 2>/dev/null || echo "already connected"

# Remove old linkpulse config if exists, then re-add
if grep -q "linkpulse" "$CONF"; then
    echo "Removing old linkpulse config..."
    sed -i '/location \/linkpulse/,/}/d' "$CONF"
    # Remove trailing empty lines before closing brace
    sed -i -e :a -e '/^\s*$/{ $d; N; ba; }' "$CONF"
fi

echo "Adding linkpulse to nginx config..."
# Remove last closing brace
sed -i '$ d' "$CONF"
cat >> "$CONF" << 'EOF'

    location /linkpulse/ {
        rewrite ^/linkpulse/(.*)$ /$1 break;
        proxy_pass http://linkpulse-app-1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
EOF
echo "nginx config updated"

docker exec cheslav-nginx nginx -t && docker restart cheslav-nginx

sleep 2
curl -sk https://localhost/linkpulse/api/health && echo ""
echo "Live: https://cheslav.space/linkpulse/"