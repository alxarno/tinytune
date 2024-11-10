import * as _ from "bootstrap"
import { lightbox  } from "./lightbox";
import { zoom } from "./zoom"
import "./search"

const onSort = (radio) => {
    Cookies.set('sort', radio.value);
    htmx.ajax('GET', window.location.pathname)
}
window.onSort = onSort
window.htmx = require('htmx.org');
window.zoom = zoom

window.onload = () => {
    document.addEventListener('htmx:afterSettle', () => {
        lightbox.init();
    })
    lightbox.init();
}