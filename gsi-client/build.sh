#!/bin/sh
set -eu

mkdir -p /out

go-winres simply \
  --icon ./icons/gsi-client.ico \
  --manifest cli \
  --product-name "Dota GSI" \
  --file-description "Dota 2 Game State Integration client" \
  --out ./cmd/client/rsrc \
  --arch amd64

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o "/out/Dota GSI.exe" ./cmd/client

build_macos_app() {
  arch="$1"
  binary_name="dota-gsi-darwin-${arch}"
  app_name="Dota GSI-${arch}.app"
  app_dir="/out/${app_name}"
  contents="${app_dir}/Contents"

  CGO_ENABLED=0 GOOS=darwin GOARCH="$arch" go build -trimpath -ldflags='-s -w' -o "/out/${binary_name}" ./cmd/client
  mkdir -p "${contents}/MacOS" "${contents}/Resources"
  cp "/out/${binary_name}" "${contents}/Resources/dota-gsi-client"
  cp ./icons/gsi-client.icns "${contents}/Resources/gsi-client.icns"
  cp ./macos/Info.plist "${contents}/Info.plist"
  cp ./macos/launcher "${contents}/MacOS/Dota GSI"
  chmod +x "${contents}/MacOS/Dota GSI" "${contents}/Resources/dota-gsi-client"
}

build_macos_app amd64
build_macos_app arm64

echo 'Compilación terminada. Revisa la carpeta output/.'
