// Via https://codepen.io/Geeyoam/pen/vLGZzG
function floodFill(newColor, x, y) {

    // Permit #FFFFFF-style color definitions
    if(newColor.search("#") >= 0) {
        newColor = hexToRGBA(newColor)
    }

    x = Math.floor(x)
    y = Math.floor(y)
    const imageData = context.getImageData(0, 0, canvas.width, canvas.height)
    const { width, height, data } = imageData
    const stack = []
    const baseColor = getColorAtPixel(imageData, x, y)
    let operator = { x, y }

    // Check if base color and new color are the same
    if (colorMatch(baseColor, newColor)) {
        return
    }

    // console.log(baseColor, newColor);

    // Add the clicked location to stack
    stack.push({ x: operator.x, y: operator.y })

    while (stack.length) {
        operator = stack.pop()
        let contiguousDown = true // Vertical is assumed to be true
        let contiguousUp = true // Vertical is assumed to be true
        let contiguousLeft = false
        let contiguousRight = false

        // Move to top most contiguousDown pixel
        while (contiguousUp && operator.y >= 0) {
            var yVal = operator.y - 1
            contiguousUp = colorMatch(getColorAtPixel(imageData, operator.x, yVal), baseColor)

            // Only decrement operator.y if we didn't roll out of bounds
            if(contiguousUp) {
                operator.y--
            }
        }

        // Move downward
        while (contiguousDown && operator.y < height) {
            setColorAtPixel(imageData, newColor, operator.x, operator.y)

            // Check left
            if (operator.x - 1 >= 0 && colorMatch(getColorAtPixel(imageData, operator.x - 1, operator.y), baseColor)) {
                if (!contiguousLeft) {
                    contiguousLeft = true
                    stack.push({ x: operator.x - 1, y: operator.y })
                }
            } else {
                contiguousLeft = false
            }

            // Check right
            if (operator.x + 1 < width && colorMatch(getColorAtPixel(imageData, operator.x + 1, operator.y), baseColor)) {
                if (!contiguousRight) {
                    stack.push({ x: operator.x + 1, y: operator.y })
                    contiguousRight = true
                }
            } else {
                contiguousRight = false
            }

            var yVal = operator.y + 1
            contiguousDown = colorMatch(getColorAtPixel(imageData, operator.x, yVal), baseColor)

            // Only increment operator.y if we didn't roll out of bounds
            if(contiguousDown) {
                operator.y++
            }
        }
    }

    context.putImageData(imageData, 0, 0)
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
    return a.r === b.r && a.g === b.g && a.b === b.b && a.a === b.a
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
