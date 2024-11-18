const gifInit = () => {
    const previews = document.querySelectorAll("img[data-animated]")
    previews.forEach(v => v.onmouseover = (event) => {
        v.setAttribute("src", v.getAttribute("src").replace("preview", "origin"))
    })
    previews.forEach(v => v.onmouseout = (event) => {
        v.setAttribute("src", v.getAttribute("src").replace("origin", "preview"))
    })
}

export {gifInit}