#!/bin/bash
set -e

cat > /opt/cheslav/nginx/conf.d/cheslav.conf << 'NGINX'
server {
    listen 80;
    listen [::]:80;
    server_name cheslav.space;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name cheslav.space;

    ssl_certificate /etc/letsencrypt/live/cheslav.space/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/cheslav.space/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;

    resolver 127.0.0.11 valid=10s;

    location / {
        root /var/www/clm;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    location /game {
        root /var/www;
        index index.html;
        try_files $uri $uri/ /game/index.html;
    }

    location /_expo/ {
        alias /var/www/game/_expo/;
    }

    location /assets/assets/ {
        alias /var/www/game/assets/assets/;
    }

    location /webrtc/speedtest {
        proxy_pass http://cheslav-speedtest:3005;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /dadata {
        proxy_pass http://cheslav-dadata:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /chainpulse {
        root /var/www;
        index index.html;
        try_files $uri $uri/ /chainpulse/index.html;
    }

    location /twa {
        root /var/www;
        index index.html;
        try_files $uri $uri/ /twa/index.html;
    }

    location /api/chainpulse/ {
        set $backend http://cheslav-chainpulse:3001;
        rewrite ^/api/chainpulse/(.*)$ /$1 break;
        proxy_pass $backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /api/patternquest/ {
        set $backend http://cheslav-patternquest:3000;
        rewrite ^/api/patternquest/(.*)$ /$1 break;
        proxy_pass $backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /socket.io/ {
        set $backend http://cheslav-patternquest:3000;
        proxy_pass $backend/socket.io/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }

    location /pharma-rag/api/ {
        proxy_pass http://pharma-rag-backend-1:8000/api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header Connection '';
        proxy_buffering off;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /pharma-rag/ {
        proxy_pass http://pharma-rag-frontend-1:80/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

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
NGINX

echo "Config written. Testing..."
docker exec cheslav-nginx nginx -t && docker restart cheslav-nginx && echo "OK!" || echo "FAILED"
sleep 2
curl -sk https://localhost/ -o /dev/null -w "Main site: HTTP %{http_code}\n"
curl -sk https://localhost/linkpulse/api/health && echo ""
curl -sk https://localhost/pharma-rag/ -o /dev/null -w "PharmaRAG: HTTP %{http_code}\n"