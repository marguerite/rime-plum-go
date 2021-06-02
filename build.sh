#!/bin/bash

mkdir -p rime-plum-go.AppDir/usr/bin
go build -o rime-plum-go.AppDir/usr/bin/rime-plum-go
wget https://github.com/AppImage/AppImageKit/releases/download/12/appimagetool-x86_64.AppImage
chmod +x appimagetool-x86_64.AppImage
wget https://github.com/AppImage/AppImageKit/releases/download/12/AppRun-x86_64
cp -r AppRun-x86_64 rime-plum-go.AppDir/AppRun
chmod +x rime-plum-go.AppDir/AppRun
cp -r rime-plum-go.desktop rime-plum-go.AppDir
wget https://github.com/rime/squirrel/raw/master/Assets.xcassets/RimeIcon.appiconset/rime-256.png
cp -r rime-256.png rime-plum-go.AppDir
./appimagetool-x86_64.AppImage rime-plum-go.AppDir
