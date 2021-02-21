# Demo: annotation with TraceOverlay

To run: 
1. Navigate to the `web` directory of this repository
1. `go build -o traceoverlay *.go cp traceoverlay ~/go/bin/` to produce a
   traceoverlay binary that should be available in your `$PATH`, assuming go is
   installed in typical fashion.
1. Change the paths in `advanced/config.json` or `basic/config.json` to
   reference this folder on your disk via absolute path (`pwd -P` will show
   you). Alternatively, navigate to this ./demo/ folder and run the program from
   here; the relative paths will then be correct.

Then in your browser, visit: http://localhost:9019

The `basic/config.json` only requires a folder with images and an output folder.
The `advanced/config.json` also includes a manifest file, so that you can
identify which images you want to annotate rather than defaulting to seeing the
whole folder.

Note: images are used with permission from the "[Pexels License](https://www.pexels.com/photo-license/)" from https://www.pexels.com/photo/christmas-cookies-on-tray-3370704/
