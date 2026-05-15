#!/bin/bash
set -e

CONF="/opt/cheslav/nginx/conf.d/cheslav.conf"

# Connect to cheslav-net
docker network connect cheslav-net linkpulse-app-1 2>/dev/null || echo "already connected"

# Remove old linkpulse config if exists
if grep -q "linkpulse" "$CONF"; then
    echo "Removing old linkpulse config..."
    # Remove all linkpulse location blocks
    python3 -c "
import re
with open('$CONF') as f:
    content = f.read()
content = re.sub(r'\n\s*location /linkpulse[^}]*\{[^}]*\}\n?', '', content)
with open('$CONF', 'w') as f:
    f.write(content)
"
fi

echo "Adding linkpulse to nginx config..."
sed -i '$ d' "$CONF"
cat >> "$CONF" << 'EOF'

    location /linkpulse/api/ {
        rewrite ^/linkpulse/api/(.*)$ /api/$1 break;
        proxy_pass http://linkpulse-app-1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

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
echo "Testing..."
curl -sk https://localhost/linkpulse/api/health && echo ""
curl -sk -o /dev/null -w "Frontend: HTTP %{http_code}\n" https://localhost/linkpulse/
echo "Live: https://cheslav.space/linkpulse/"