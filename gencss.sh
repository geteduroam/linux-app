#!/bin/sh

# get the base64 for the heart
export IMG_HEART=$(cat ./cmd/geteduroam-gui/resources/images/heart.svg | base64 -w 0)
export IMG_LOGO=$(cat ./cmd/geteduroam-gui/resources/images/logo.png | base64 -w 0)
envsubst < cmd/geteduroam-gui/resources/window.css.template > cmd/geteduroam-gui/resources/window.css
