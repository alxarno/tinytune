import videojs from 'video.js';

let initializedVideos = [];
export const videoClear = () => {
    initializedVideos = []
}
export const videoInit = (element) => {
    let src = element.getAttribute("src");
    if(initializedVideos.includes(src)) return;
    initializedVideos.push(src);
    const origin = document.querySelector(`a[href="${src.includes("rts") ? src.replace("rts", "origin") : src}"]`)
    const wrapper = document.createElement("div")
    element.parentElement.appendChild(wrapper)

    element.remove()
    const videoElement = document.createElement("video")
    videoElement.classList.add("video-js");

    if (src[0] == "#") {
        src = src.substring(1)
    }
    videoElement.innerHTML = `<source src="${src}" type="${origin.getAttribute("data-extension")}">`
    wrapper.appendChild(videoElement)

    const videoJSOptions = {
        controls: true,
        autoplay: false,
        preload: 'metadata',
        enableSmoothSeeking: true,
        experimentalSvgIcons: true,
    }
    var player = videojs(videoElement, videoJSOptions)

    player.ready(() => {
        videoElement.parentElement.onpointerdown = (e) => {
            e.stopImmediatePropagation();
        }
        const parent = videoElement.parentElement
        parent.style.width = origin.getAttribute("data-width") + "px";
        parent.style.height = origin.getAttribute("data-height") + "px";

        const lightboxWrapper = wrapper.parentElement.parentElement

        var observer = new MutationObserver(function(mutations) {
            mutations.forEach(function(mutationRecord) {
                const transform = Number(lightboxWrapper.style.transform.replace("translateX(", "").replace("px)", ""))
                if(transform != 0) {
                    player.pause()
                }
            });    
        });
        observer.observe(lightboxWrapper, { attributes : true, attributeFilter : ['style'] });
    })
}