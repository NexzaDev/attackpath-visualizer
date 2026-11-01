package assets

import "embed"

//go:embed icons/icon.svg icons/icon.png icons/icon.ico icons/icon.icns icons/icon-*.png icons/nav/*.svg fonts/*.ttf fonts/OFL.txt docs/README.md docs/USAGE.md docs/LICENSE.md docs/LICENSES.md docs/THIRD_PARTY_NOTICES.md all:licenses demo.json linux/attackpath-visualizer.desktop
var Files embed.FS
