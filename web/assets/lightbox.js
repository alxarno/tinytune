import PhotoSwipeLightbox from '/static/lightbox/photoswipe-lightbox.esm.min.js';
import PhotoSwipe from '/static/lightbox/photoswipe.esm.min.js';
const lightbox = new PhotoSwipeLightbox({
    gallery: '#gallery--getting-started',
    children: 'li > .image-lightbox',
    pswpModule: PhotoSwipe
});
lightbox.init();
lightbox.on('change', () => {
    // triggers when slide is switched, and at initialization
    const currentUrl = lightbox.pswp.currSlide.data.src;
    const newUrl = `${window.location.origin}${window.location.pathname}#?item=${currentUrl.split("origin/")[1]}`
    history.pushState({}, "", currentUrl);
    history.pushState({}, "", newUrl);
});
lightbox.on('close', () => {
    history.pushState({}, "", `${window.location.origin}${window.location.pathname}`);
});

window.onload = () => {
    document.addEventListener('htmx:afterSettle', () => {
        lightbox.init();
    })
}