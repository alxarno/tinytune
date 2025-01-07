import Hls from 'hls.js';

export const videoInit = (video) => {
    const originSrc = video.getAttribute("src").replace("hls", "origin").replace(".m3u8", "");
    const origin = document.querySelector(`a[href="${originSrc}"]`)
    video.style.width = origin.getAttribute("data-width") + "px";
    video.style.height = origin.getAttribute("data-height") + "px";
    video.style.display = "block";
    video.style.objectFit = "cover";
    video.setAttribute("poster", originSrc.replace("origin", "preview"))
    

    if (Hls.isSupported()) {
        var hls = new Hls();
        hls.loadSource(video.getAttribute("src"));
        hls.attachMedia(video);
        hls.on(Hls.Events.BUFFER_CREATED, () => video.removeAttribute("poster"));
    }
}

export const initVideoTransitions = (video) => {
    video.parentElement.onpointerdown = (e) => {
        e.stopImmediatePropagation();
    }

    const lightboxWrapper = video.parentElement.parentElement

    var observer = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutationRecord) {
            const transform = Number(lightboxWrapper.style.transform.replace("translateX(", "").replace("px)", ""))
            if(transform != 0) {
                video.pause()
            }
        });    
    });
    observer.observe(lightboxWrapper, { attributes : true, attributeFilter : ['style'] });
}