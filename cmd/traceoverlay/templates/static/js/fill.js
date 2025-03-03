// Via https://codepen.io/Geeyoam/pen/vLGZzG
function floodFill(newColor, x, y) {
    // Permit #FFFFFF-style color definitions
    if (newColor.search("#") >= 0) {
        newColor = hexToRGBA(newColor);
    }

    x = Math.floor(x);
    y = Math.floor(y);
    const imageData = context.getImageData(0, 0, canvas.width, canvas.height);
    const { width, height } = imageData;
    const baseColor = getColorAtPixel(imageData, x, y);

    // Create an array to track visited pixels.
    const visited = new Array(width * height).fill(false);
    
    // Helper function to compute index in the visited array.
    function getIndex(x, y) {
        return y * width + x;
    }

    const stack = [];
    stack.push({ x, y });

    while (stack.length) {
        const { x: currentX, y: currentY } = stack.pop();

        // Skip if out of bounds.
        if (currentX < 0 || currentX >= width || currentY < 0 || currentY >= height) {
            continue;
        }

        const idx = getIndex(currentX, currentY);
        if (visited[idx]) {
            continue;
        }
        visited[idx] = true;

        // For debugging: log the current pixel.
        // console.log("Filling pixel:", currentX, currentY);

        // If the pixel doesn't match the baseColor, skip it.
        if (!colorMatch(getColorAtPixel(imageData, currentX, currentY), baseColor)) {
            continue;
        }

        // Fill the current pixel.
        setColorAtPixel(imageData, newColor, currentX, currentY);

        // Add neighbors to the stack.
        stack.push({ x: currentX + 1, y: currentY });
        stack.push({ x: currentX - 1, y: currentY });
        stack.push({ x: currentX, y: currentY + 1 });
        stack.push({ x: currentX, y: currentY - 1 });
    }

    context.putImageData(imageData, 0, 0);
}

function getColorAtPixel(imageData, x, y) {
    const { width, data } = imageData

    // console.log("Requesting", x, y);

    return {
        r: data[4 * (width * y + x) + 0],
        g: data[4 * (width * y + x) + 1],
        b: data[4 * (width * y + x) + 2],
        a: data[4 * (width * y + x) + 3]
    }
}

function setColorAtPixel(imageData, color, x, y) {
    const { width, data } = imageData

    // console.log(color)
    // console.log(color.r)
    // console.log(color.a)
    // console.log(color.a & 0xff)

    data[4 * (width * y + x) + 0] = color.r & 0xff
    data[4 * (width * y + x) + 1] = color.g & 0xff
    data[4 * (width * y + x) + 2] = color.b & 0xff
    data[4 * (width * y + x) + 3] = color.a & 0xff
}

function colorMatch(a, b) {
    // return a.r === b.r && a.g === b.g && a.b === b.b && a.a === b.a

    // Ignore alpha, just compare r, b, and g channels
    return a.r === b.r && a.g === b.g && a.b === b.b
}

function hexToRGBA(hexStr) {
    let sanitized = hexStr.replace("#", "")

    return {
        r: parseInt(sanitized.substring(0,2), 16),
        g: parseInt(sanitized.substring(2,4), 16),
        b: parseInt(sanitized.substring(4,6), 16),
        a: previewAlpha, // previewAlpha from traceoverlay.js
    }
}

// Directly from https://stackoverflow.com/a/3627747/199475
function rgb2hex(rgb) {
    if (/^#[0-9A-F]{6}$/i.test(rgb)) return rgb;

    rgb = rgb.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/);
    function hex(x) {
        return ("0" + parseInt(x).toString(16)).slice(-2);
    }
    return "#" + hex(rgb[1]) + hex(rgb[2]) + hex(rgb[3]);
}
