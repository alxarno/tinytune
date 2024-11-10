import PhotoSwipeLightbox from 'photoswipe/lightbox';
import PhotoSwipe from "photoswipe"
import videojs from 'video.js';

const lightbox = new PhotoSwipeLightbox({
    gallery: '#gallery--getting-started',
    children: 'li > .image-lightbox',
    pswpModule: PhotoSwipe,
});
const videoInit = () => {
    const videoJSOptions = {
        controls: true,
        autoplay: false,
        preload: 'metadata',
        enableSmoothSeeking: true,
        experimentalSvgIcons: true,
    }
    if (document.querySelector('.video-js')) {
        videojs(document.querySelector('.video-js'), videoJSOptions);
    }
}
lightbox.on('change', () => {
    videoInit()
});

lightbox.on('bindEvents', () => {
    videoInit()
});
lightbox.addFilter('itemData', (itemData, index) => {
    if (itemData.element.dataset.pswpType == "video") {
        itemData.src = itemData.element.getAttribute("href");
        itemData.width = itemData.element.dataset.pswpWidth;
        itemData.height = itemData.element.dataset.pswpHeight;
        itemData.srcType = itemData.element.dataset.pswpSourceType;
    }
    return itemData
});
lightbox.on('contentLoad', (e) => {
    const { content } = e;
    if (content.type === 'video') {
        e.preventDefault();

        content.element = document.createElement('div');
        content.element.className = 'pswp__video-container';

        const video = document.createElement('video');
        video.setAttribute("class", "video-js vjs-default-skin")
        video.setAttribute("width", content.data.width)
        video.setAttribute("height", content.data.height)

        const source = document.createElement("source")
        source.setAttribute("src", content.data.src)
        source.setAttribute("type", content.data.srcType)
        video.appendChild(source)

        content.element.appendChild(video);
    }
});

export {lightbox}