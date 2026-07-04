// Package web embeds the built frontend assets (web/dist) into the binary.
// Rebuild with `npm run build` inside web/ after frontend changes; the
// resulting dist/ is what gets baked in at compile time via go:embed.
package web

import "embed"

//go:embed all:dist
var Assets embed.FS
