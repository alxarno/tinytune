import { Popover } from 'bootstrap';
require("fslightbox")
import { videoInit, initVideoTransitions  } from "./video";
import { zoom } from "./zoom"
import { observeDOM } from "./dom";
import { onSearch, highlightSearchResults } from "./search"
import { onButtonUpClick, buttonUpInit } from "./button-up"
import { gifInit } from "./gif"
import Cookies from "js-cookie"
import htmx from "htmx.org"

const onSort = (radio) => {
    Cookies.set('sort', radio.value);
    htmx.ajax('GET', window.location.pathname)
}
window.onSort = onSort
window.htmx = require('htmx.org');
window.zoom = zoom
window.onSearch = onSearch
window.onButtonUpClick = onButtonUpClick


const initLightBox = () => {
    fsLightbox = new FsLightbox();
    const previews = Array.from(document.getElementsByClassName("image-lightbox"))
    previews.forEach((v, i) => v.setAttribute("index", i))
    previews.forEach((v, i) => v.onclick = (e) => {
        e.preventDefault();
        fsLightbox.open(i);
    });
    fsLightbox.props.sources = previews.map(v => v.getAttribute("data-src") ? v.getAttribute("data-src") : v.getAttribute("href"))
    fsLightbox.props.types = previews.map(v => v.getAttribute("type"));
    fsLightbox.props.thumbs = previews.map(v => v.getAttribute("href").replace("origin", "preview"));
    fsLightbox.props.loadOnlyCurrentSource = true;
    fsLightbox.props.onOpen = () => {
        const videosItems = Array.from(document.querySelectorAll("video"))
        videosItems.forEach(initVideoTransitions)
        const streams = videosItems.filter(e => e.getAttribute("src").includes("hls"))
        streams.forEach(videoInit)
    }
}

const onPopState = () => {
    initLightBox();
    gifInit();
    buttonUpInit();
}

window.onload = () => {
    window.addEventListener('popstate', onPopState)
    document.addEventListener('htmx:afterSettle', () => {
        initLightBox();
        gifInit();
        buttonUpInit();
    })
    initLightBox();
    highlightSearchResults();
    buttonUpInit();
    gifInit();
}
