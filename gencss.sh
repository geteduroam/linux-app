#!/bin/sh

export IMG_HEART=$(cat ./cmd/geteduroam-gui/resources/images/heart.svg | base64 -w 0)
# geteduroam
export IMG_LOGO=$(cat ./cmd/geteduroam-gui/resources/images/geteduroam.png | base64 -w 0)
envsubst < cmd/geteduroam-gui/resources/window.css.template > cmd/geteduroam-gui/resources/window_geteduroam.css
# getgovroam
export IMG_LOGO=$(cat ./cmd/geteduroam-gui/resources/images/getgovroam.png | base64 -w 0)
envsubst < cmd/geteduroam-gui/resources/window.css.template > cmd/geteduroam-gui/resources/window_getgovroam.css
