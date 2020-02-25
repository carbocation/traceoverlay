// Inspiration for path drawing rather than pixel shading comes from
// http://jsfiddle.net/m1erickson/AEYYq/ , as does the code for undo.

var canvas = document.getElementById("imgCanvas");
var context = canvas.getContext("2d");
context.globalAlpha = 1.0;
context.imageSmoothingEnabled = false;
// context.fillStyle = "#ff0000"; //"rgba(255, 0, 0, 1)";
// context.fillStyle = "#ff0000";
// context.strokeStyle = "#ff0000";
var previewAlpha = 160;
var saveAlpha = 255;

// Brush variables
var brush = "stroke";
var brushSize = 1;
var brushColor = "#ff0000"; //"rgba(255, 0, 0, 1)"; //"#FF0000"; //{r:0xff, g:0x00, b:0x00, a:0xff}; //"#FF0000";

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

function setBrushColor(color) {
    if(color == "") {

        brushColor = "rgba(0, 0, 0, 1)";
        // setBrush("eraser");
        return
    }

    brushColor = color;
    if(brush == "fill" || brush == "eraser") {
        setBrush("stroke");
    }
}

function setBrushSize(size) {
    if(brush == "fill") {
        brush = "stroke";
    }

    if(size < 1) {
        size = 1;
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

    if(brush == "ekg") {
        setBrushSize(canvas.height);
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

    var pos = getMousePos(canvas, e);

    if(brush == "line") {
        // If drawing a straight line segment, we aren't drawing along the way,
        // so we need to commit our stroke here:
        context.fillStyle = brushColor;
        context.strokeStyle = brushColor;
        context.lineTo(pos.x, pos.y);
        context.stroke();
    }

    fullyShade(previewAlpha);

    thisY = pos.y;
    if(brush == "ekg") {
        thisY = canvas.height / 2;
    }

    points.push({
        x: pos.x,
        y: thisY,
        brush: brush,
        size: brushSize,
        color: brushColor,
        mode: "end"
    });
    lastX = pos.x;
    lastY = thisY;

    mouseDown = false;
}

// Initiate a new line segment or flood fill
function start(e) {
    var pos = getMousePos(canvas, e);

    thisY = pos.y;
    if(brush == "ekg") {
        thisY = canvas.height / 2;
    }

    lastX = pos.x;
    lastY = thisY;

    points.push({
        x: pos.x,
        y: thisY,
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
        // floodFill({r: 0x00, g: 0x00, b: 0xff, a: 0xff}, pos.x, pos.y);
        floodFill(brushColor, pos.x, pos.y);

        fullyShade(previewAlpha);

        mouseDown = false;

        return false
    }

    context.beginPath();
    // context.lineJoin = 'round';
    // context.lineCap = 'round';
    
    // Ensure that brush vars are set
    if(context.lineWidth != brushSize) {
        context.lineWidth = brushSize;
    }

    context.moveTo(pos.x, thisY);

    mouseDown = true;
}

// Continue drawing a line segment
function draw(e) {
    if(!mouseDown) {
        return
    }

    if(brush == "line") {
        // Don't keep the mouse down for a straight line segment
        return
    }

    var pos = getMousePos(canvas, e);

    context.fillStyle = brushColor;
    context.strokeStyle = brushColor;

    thisY = pos.y;
    if(brush == "ekg") {
        thisY = canvas.height / 2;
    }

    context.lineTo(pos.x, thisY);
    context.stroke();

    // command pattern stuff
    points.push({
        x: pos.x,
        y: thisY,
        brush: brush,
        size: brushSize,
        color: brushColor,
        mode: "draw"
    });
    lastX = pos.x;
    lastY = thisY;
}

canvas.addEventListener('mousemove', draw, false);
canvas.addEventListener('mouseout', stop, false);
canvas.addEventListener('mouseup', stop, false);
canvas.addEventListener('mousedown', start, false);

// Support touch devices. Actually, since I'm referring to mouse events
// throughout the functions in this script, it is probably easiest to creat a
// mouse event from each touch event. See
// http://bencentra.com/code/2014/12/05/html5-canvas-touch-events.html
function getTouchPos(canvasDom, touchEvent) {
    var rect = canvasDom.getBoundingClientRect();
    return {
        x: touchEvent.touches[0].clientX - rect.left,
        y: touchEvent.touches[0].clientY - rect.top
    };
}
canvas.addEventListener('touchmove', function (e) {
    var touch = e.touches[0];
    var mouseEvent = new MouseEvent("mousemove", {
        clientX: touch.clientX,
        clientY: touch.clientY
    });
    canvas.dispatchEvent(mouseEvent);
}, false);
canvas.addEventListener('touchcancel', function (e) {
    var mouseEvent = new MouseEvent("mouseup", {});
    canvas.dispatchEvent(mouseEvent);
}, false);
canvas.addEventListener('touchend', function (e) {
    var mouseEvent = new MouseEvent("mouseup", {});
    canvas.dispatchEvent(mouseEvent);
}, false);
canvas.addEventListener('touchstart', function (e) {
    mousePos = getTouchPos(canvas, e);
    var touch = e.touches[0];
    var mouseEvent = new MouseEvent("mousedown", {
        clientX: touch.clientX,
        clientY: touch.clientY
    });
    canvas.dispatchEvent(mouseEvent);
}, false);


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

    // Clear the contents of the canvas
    context.clearRect(0, 0, canvas.width, canvas.height);

    // Add back the original image, if it exists:
    context.globalCompositeOperation="source-over";
    context.drawImage(preExistingImage, 0, 0);

    // If no points have been created yet, we're done:
    if (points.length == 0) {
        return;
    }

    // console.log(points);

    for (var i = 0; i < points.length; i++) {

        var pt = points[i];
        // console.log(pt.color);

        if(pt.brush == "eraser") {
            context.globalCompositeOperation="destination-out";
        } else if(pt.brush == "fill") {
            context.globalCompositeOperation="source-over";
            floodFill(pt.color, pt.x, pt.y);
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

// Support touch devices
document.getElementById("undo").addEventListener('touchend', undoStop, false);
document.getElementById("undo").addEventListener('touchcancel', undoStop, false);
document.getElementById("undo").addEventListener('touchstart', undoStart, false);

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

// Switch with keys
$(document).on("keypress", function(event){
    event.preventDefault();
    // console.log(event); 
    if(event.key == "e"){
        setBrush('eraser');
        flashMessage("Eraser mode" + " (key " + event.key + ")");
    } else if(event.key == "s"){
        setBrush('stroke');
        flashMessage("Brush: " + brush + " (key " + event.key + ")");
    } else if(event.key == "k"){
        setBrush('ekg');
        flashMessage("Brush: " + brush + " (key " + event.key + ")");
    } else if(event.key == "f"){
        setBrush('fill');
        flashMessage("Brush: " + brush + " (key " + event.key + ")");
    } else if(event.key == "l") {
        setBrush('line');
        flashMessage("Brush: " + brush + " (key " + event.key + ")");
    } else if(event.key == "n") {
        saveCanvas();
    } else if(event.key == "z") {
        setBrushSize(brushSize - 1);
        flashMessage("Brush size now " + brushSize + " (key " + event.key + ")");
    } else if(event.key == "x") {
        setBrushSize(brushSize + 1);
        flashMessage("Brush size now " + brushSize + " (key " + event.key + ")");
    } else if(event.key == "q") {
        newBrushName = prevBrush();
        flashMessage("Brush: " + newBrushName + " (key " + event.key + ")");
    } else if(event.key == "w") {
        newBrushName = nextBrush();
        flashMessage("Brush: " + newBrushName + " (key " + event.key + ")");
    } else if(event.key == "h") {
        msg = toggleVisibility();
        flashMessage(msg + " canvas" + " (key " + event.key + ")");
    }
});

var canvasVisibility = "visible";
function toggleVisibility() {
    if(canvasVisibility == "hidden"){
        canvasVisibility = "visible";
        redrawAll();
        fullyShade(previewAlpha);

        return "Revealing"
    } else {
        canvasVisibility = "hidden";
        context.clearRect(0, 0, canvas.width, canvas.height);

        return "Hiding"
    }
}

var flashTimeout;
function flashMessage(message) {
    // If the timeout is already set for a prior message, block that from
    // prematurely hiding the new one
    clearTimeout(flashTimeout);

    // Fetch the message box and make it visible
    var target = document.getElementById("drawmessage");
    target.style.visibility = "visible";
    target.style.border = "4px solid " + brushColor;

    // Update its contents
    target.textContent = message;

    // Set a timeout again
    flashTimeout = setTimeout(function(){
        target.style.visibility = "hidden";
    }, 350);
}

function nextBrush() {
    return changeBrush("next");
}

function prevBrush() {
    return changeBrush("prev");
}

function changeBrush(dir) {    
    var labels = document.getElementById("labels");
    if(labels.childElementCount < 1){
        return ""
    }

    var activeIndex = 0
    Array.from(labels.children).forEach(function(item, i, array){
        bgCol = window.getComputedStyle(item, null).getPropertyValue('background-color');

        if(rgb2hex(bgCol) == brushColor){
            activeIndex = i;
        }
    })

    if(dir == "prev"){
        desiredIndex = activeIndex - 1;
    } else {
        desiredIndex = activeIndex + 1;
    }

    if( desiredIndex > Array.from(labels.children).length - 1) {
        desiredIndex = Array.from(labels.children).length - 1;
    } else if( desiredIndex < 0) {
        desiredIndex = 0;
    }

    // This is the label we want
    elem = Array.from(labels.children)[desiredIndex];

    // Special case: exit early if it's the background color
    // console.log("'" + elem.style.backgroundColor + "'");
    if(elem.style.backgroundColor == ""){
        // console.log("Setting to blank?")
        setBrushColor("");
        return elem.textContent;
    }

    // General case: get the computed background color and use that.
    setBrushColor(rgb2hex(window.getComputedStyle(elem, null).getPropertyValue('background-color')));

    return elem.textContent;
}
