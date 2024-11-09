import * as _ from "bootstrap"
import "./img-preview";
import { lightbox  } from "./lightbox";
import "./search"
import "./zoom"

const onSort = (radio) => {
    Cookies.set('sort', radio.value);
    htmx.ajax('GET', window.location.pathname)
}
window.onSort = onSort
window.htmx = require('htmx.org');

window.onload = () => {
    document.addEventListener('htmx:afterSettle', () => {
        lightbox.init();
    })
    lightbox.init();
}