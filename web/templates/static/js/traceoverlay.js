// Inspiration for path drawing rather than pixel shading comes from
// http://jsfiddle.net/m1erickson/AEYYq/ , as does the code for undo.

var canvas = document.getElementById("imgCanvas");
var context = canvas.getContext("2d");
context.globalAlpha = 1.0;
context.imageSmoothingEnabled = false;
context.fillStyle = "rgba(255, 0, 0, 1)";
// context.fillStyle = "#ff0000";
// context.strokeStyle = "#ff0000";
var previewAlpha = 196;
var saveAlpha = 255;

// Brush variables
var brush = "stroke";
var brushSize = 1;
var brushColor = "rgba(255, 0, 0, 1)"; //"#FF0000"; //{r:0xff, g:0x00, b:0x00, a:0xff}; //"#FF0000";

// Undo
var lastX;
var lastY;
var points = [];

// Helpers
function getMousePos(canvas, evt) {
    const newLocal = canvas.getBoundingClientRect();
    var rect = newLocal;
    return {
        x: (evt.clientX - rect.left) / (rect.right - rect.left) * canvas.width,
        y: (evt.clientY - rect.top) / (rect.bottom - rect.top) * canvas.height
    };
}

function setBrushSize(size) {
    if(brush == "fill") {
        brush = "stroke";
    }

    brushSize = size;
}

function setBrush(newBrush) {
    brush = newBrush;

    if(brush == "eraser") {
        // Via https://stackoverflow.com/a/25916334/199475
        context.globalCompositeOperation="destination-out";
        console.log("Creating transparent brush")
        // context.fillStyle = "rgba(0, 0, 0, 0)";
    } else {
        context.globalCompositeOperation="source-over";
        console.log("Creating opaque brush")
        // context.fillStyle = "rgba(0, 0, 0, 1)";
    }
}

// Handlers for the position of the mouse, and whether or not the mouse button
// has been held down.
var mouseDown = false;

// Complete this line segment
function stop(e) {
    if(!mouseDown) {
        return
    }
    fullyShade(previewAlpha);

    var pos = getMousePos(canvas, e);

    points.push({
        x: pos.x,
        y: pos.y,
        brush: brush,
        size: brushSize,
        color: brushColor,
        mode: "end"
    });
    lastX = pos.x;
    lastY = pos.y;

    mouseDown = false;
}

// Initiate a new line segment or flood fill
function start(e) {
    var pos = getMousePos(canvas, e);

    lastX = pos.x;
    lastY = pos.y;

    points.push({
        x: pos.x,
        y: pos.y,
        brush: brush,
        size: brushSize,
        color: brushColor,
        mode: "begin"
    });

    if(brush == "fill") {
        console.log("Fill");

        const newLocal = canvas.getBoundingClientRect();
        var rect = newLocal;

        // TODO: Figure out how to use the same color object for fill and draw
        floodFill({r: 0x00, g: 0x00, b: 0xff, a: 0xff}, pos.x, pos.y);
        // floodFill(brushColor, pos.x, pos.y);

        fullyShade(previewAlpha);

        mouseDown = false;

        return false
    }

    context.beginPath();
    
    // Ensure that brush vars are set
    if(context.lineWidth != brushSize) {
        context.lineWidth = brushSize;
    }

    context.moveTo(pos.x, pos.y);

    mouseDown = true;
}

// Continue drawing a line segment
function draw(e) {
    if(!mouseDown) {
        return
    }

    var pos = getMousePos(canvas, e);

    context.fillStyle = brushColor;
    context.strokeStyle = brushColor;

    context.lineTo(pos.x, pos.y);
    context.stroke();

    // command pattern stuff
    points.push({
        x: pos.x,
        y: pos.y,
        brush: brush,
        size: brushSize,
        color: brushColor,
        mode: "draw"
    });
    lastX = pos.x;
    lastY = pos.y;
}

canvas.addEventListener('mousemove', draw, false);
canvas.addEventListener('mouseout', stop, false);
canvas.addEventListener('mouseup', stop, false);
canvas.addEventListener('mousedown', start, false);

function saveCanvas() {
    // context.globalAlpha = 1.0;
    fullyShade(saveAlpha);

    var base64Image = CanvasToBMP.toDataURL(document.getElementById('imgCanvas'));

    document.getElementById("imgBase64").value = base64Image;
    document.getElementById("saveImage").submit();

    // Reset our alpha
    fullyShade(previewAlpha);
}

function ajaxSaveCanvas(imageIndex) {
    var base64Image = CanvasToBMP.toDataURL(document.getElementById('imgCanvas'));

    $.ajax({
        type: "POST",
        url: "/traceoverlay/" + imageIndex,
        data: { 
            imgBase64: base64Image
        }
    }).done(function(o) {
        console.log(base64Image);
        console.log('saved'); 
    });
}

function downloadCanvas() {

    var link = document.createElement('a');
    link.download = 'canvas.bmp';
    link.href = CanvasToBMP.toBlob(document.getElementById('imgCanvas'));
    link.click();

    return
}

// shadeAlpha is an int from 0-255
function fullyShade(shadeAlpha) {
    var imageData = context.getImageData(0,0,canvas.width, canvas.height);
    var pixels = imageData.data;
    var numPixels = pixels.length;

    context.clearRect(0, 0, canvas.width, canvas.height);

    for (var i = 0; i < numPixels; i++) {
        if (pixels[i*4+3] <= 32) {
            pixels[i*4+3] = 0;
        } else {
            pixels[i*4+3] = shadeAlpha;
        }
    }
    context.putImageData(imageData, 0, 0);
}

function redrawAll() {

    // console.log("Redrawing " + points.length + " points");
    // console.log(points[0]);
    // console.log(points[1]);

    if (points.length == 0) {
        return;
    }

    context.clearRect(0, 0, canvas.width, canvas.height);

    // console.log(context);

    for (var i = 0; i < points.length; i++) {

        var pt = points[i];

        if(pt.brush == "eraser") {
            context.globalCompositeOperation="destination-out";
        } else if(pt.brush == "fill") {
            context.globalCompositeOperation="source-over";
            // TODO: Figure out how to use the same color object for fill and draw
            floodFill({r: 0x00, g: 0xff, b: 0x00, a: 0xff}, pt.x, pt.y);
            continue;
        } else {
            context.globalCompositeOperation="source-over";
        }

        if (context.lineWidth != pt.size) {
            context.lineWidth = pt.size;
        }
        if (context.strokeStyle != pt.color || context.fillStyle != pt.color) {
            context.strokeStyle = pt.color;
            context.fillStyle = pt.color;
        }
        if (pt.mode == "begin") { // || begin) {
            context.beginPath();
            context.moveTo(pt.x, pt.y);
        }
        context.lineTo(pt.x, pt.y);
        context.stroke();
    }

    // console.log("Finished re-drawing");
}

var interval;
document.getElementById("undo").addEventListener('mouseout', undoStop, false);
document.getElementById("undo").addEventListener('mouseup', undoStop, false);
document.getElementById("undo").addEventListener('mousedown', undoStart, false);

function undoStop() {
    clearInterval(interval);
    fullyShade(previewAlpha);
}

function undoStart() {
    interval = setInterval(undoLast, 20);
}

function undoLast() {
    points.pop();
    redrawAll();
}