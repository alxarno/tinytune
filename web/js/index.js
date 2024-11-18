import { Popover } from 'bootstrap';
require("fslightbox")
import { videoInit, videoClear  } from "./video";
import { zoom } from "./zoom"
import { observeDOM } from "./dom";
import { onSearch, highlightSearchResults } from "./search"
import { onButtonUpClick, buttonUpInit } from "./button-up"
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


let fsLightbox = new FsLightbox();

const originOpen = (index) => {
    fsLightbox.open(index)
}

const initLightBox = () => {
    fsLightbox = new FsLightbox();
    const previews = Array.from(document.getElementsByClassName("image-lightbox"))
    previews.forEach((v, i) => v.setAttribute("index", i))
    previews.forEach((v, i) => v.onclick = (e) => {e.preventDefault(); originOpen(i)});
    fsLightbox.props.sources = previews.map(v => v.getAttribute("href"))
    fsLightbox.props.types = previews.map(v => v.getAttribute("type"));
    fsLightbox.props.loadOnlyCurrentSource = true;
    fsLightbox.props.onOpen = () => {
        const videosItems = Array.from(document.querySelectorAll("video.fslightbox-source"))
        videosItems.forEach(videoInit)
        return
    }
    fsLightbox.props.onClose = () => {
        videoClear();
    }
}

window.onload = () => {
    document.addEventListener('htmx:afterSettle', () => {
        initLightBox();
    })
    initLightBox();
    highlightSearchResults();
    buttonUpInit();
    observeDOM()(document.body, (m) => {
        let addedNodes = []
        m.forEach(record => record.addedNodes.length & addedNodes.push(...record.addedNodes));
        let videoNodes = addedNodes.filter(v => v.tagName == "VIDEO" && v.classList && v.classList.contains("fslightbox-source"))
        if(!videoNodes.length) return;
        videoNodes.forEach((v) => videoInit(v))
    })
}
