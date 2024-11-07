const imageOnload = (id) => {
    const img = document.getElementById(`video-${id}`);
    if(!img) return;
    const wrap = img.parentElement;
    const oneImageHeightFromThumbnail = img.naturalHeight / 5;
    const widthRatio = wrap.clientWidth / img.naturalWidth;
    let imgWidth = wrap.clientWidth
    let imgHeight = oneImageHeightFromThumbnail * widthRatio;
    if(imgHeight > wrap.clientHeight) {
        imgWidth = imgWidth * (wrap.clientHeight/imgHeight)
        imgHeight = wrap.clientHeight

    }
    img.style.height = imgHeight+"px";
    img.style.width = imgWidth+"px";
}