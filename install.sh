#!/usr/bin/env bash

echo "Creating system files..."
mkdir -p /usr/share/thermoPi/dist


echo "Compiling Vue.js webapp..."
cd public
echo "Installing packages..."
npm install
echo "Building Vue.js project..."
npm run build
echo "Moving Vue.js project to /usr/share/thermoPi/dist"
mv dist/* /usr/share/thermoPi/dist/
cd ..

echo "Installing Golang backend"
go install -i /usr/local/bin
echo "thermoPi has been installed to /usr/local/bin"
