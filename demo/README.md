# Demo: annotation with TraceOverlay

To run: 
1. Change the absolute paths in `advanced/config.json` or `basic/config.json` to
   reference this folder on your disk (`pwd -P` will show you)
1. Navigate to the `web` directory of this repository
1. `go build -o traceoverlay.linux *.go && ./traceoverlay.linux -config ../demo/config.json`

Then in your browser, visit: http://localhost:9019

The `basic/config.json` only requires a folder with images and an output folder.
The `advanced/config.json` also includes a manifest file, so that you can
identify which images you want to annotate rather than defaulting to seeing the
whole folder.

Note: images are used with permission from the "[Pexels License](https://www.pexels.com/photo-license/)" from https://www.pexels.com/photo/christmas-cookies-on-tray-3370704/
