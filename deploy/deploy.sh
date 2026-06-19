#!/bin/bash
set -e

DEPLOY_DIR="/var/www/garagebrain"

echo "Building..."
cd apps/garagebrain/backend && go build -o bin/server ./cmd/server
cd ../../gateway && go build -o bin/gateway ./cmd/server
cd ../garagebrain/frontend && npm run build
cd ../../..

echo "Deploying to $DEPLOY_DIR..."
sudo mkdir -p $DEPLOY_DIR/bin $DEPLOY_DIR/frontend/dist $DEPLOY_DIR/templates

sudo cp apps/garagebrain/backend/bin/server $DEPLOY_DIR/bin/
sudo cp apps/gateway/bin/gateway $DEPLOY_DIR/bin/
sudo cp -r apps/garagebrain/frontend/dist/* $DEPLOY_DIR/frontend/dist/
sudo cp apps/garagebrain/backend/templates/passport.html $DEPLOY_DIR/templates/
sudo cp apps/garagebrain/backend/.env $DEPLOY_DIR/.env

sudo chown -R www-data:www-data $DEPLOY_DIR

echo "Copying configs..."
sudo cp deploy/systemd/garagebrain-api.service /etc/systemd/system/
sudo cp deploy/systemd/garagebrain-gateway.service /etc/systemd/system/
sudo cp deploy/nginx/garagebrain.conf /etc/nginx/sites-available/garagebrain

echo "Enabling nginx..."
sudo ln -sf /etc/nginx/sites-available/garagebrain /etc/nginx/sites-enabled/garagebrain

echo "Restarting services..."
sudo systemctl daemon-reload
sudo systemctl restart garagebrain-gateway
sudo systemctl restart garagebrain-api
sudo nginx -t && sudo systemctl reload nginx

echo "Deploy complete!"
