#!/usr/bin/env bash

# TODO: Add option to install any git repo with an index.html (or universal build system).
install_vue() {
    echo "Cloning git repo..."
    git clone https://github.com/christopherm99/thermopi-webapp public
    cd public
    echo "Installing production dependencies..."
    npm install --production
    echo "Building Vue.js project..."
    npm run build
    echo "Moving Vue.js project to /usr/share/thermoPi/dist..."
    mv dist/* /usr/share/thermoPi/dist/
    cd ..
}

echo "Creating program files..."
mkdir -p /usr/share/thermoPi/dist

read -p "Do you wish to install the default webapp, https://github.com/christopherm99/thermopi-webapp [Y/n]?" yn
case ${yn} in
    [Nn]* ) echo "Not installing default webapp. To host files, place them in /usr/share/thermoPi/dist";;
    * ) install_vue;;
esac

echo "Installing backend dependencies"
dep ensure

echo "Installing Golang backend..."
go install -i /usr/local/bin

echo "Setting up program files..."
cp schedule.csv /usr/share/thermoPi/schedule.csv
chown -r ${USER} /usr/share/thermoPi/dist

echo "thermoPi has been installed to /usr/local/bin"
