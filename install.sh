#!/usr/bin/env bash

echo "Compiling Vue.js webapp..."
cd public
echo "Installing packages..."
npm install
echo "Building Vue project..."
npm run build
cd ..

echo "Installing Golang backend"
go install -i
if [[ -z $GOBIN ]]
then
echo "GOBIN unset. Is go installed properly?"
echo "Nonetheless, thermoPi has been likely installed to GOPATH/bin/${PWD##*/}"
echo "Run go env to discover your GOPATH."
else
echo "thermoPi has been installed to $GOBIN${PWD##*/}"
fi
