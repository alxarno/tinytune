const imageOnload = (id) => {
    const img = document.getElementById(`video-${id}`);
    const wrap = img.parentElement;
    const oneImageHeightFromThumbnail = img.naturalHeight / 5;
    const widthRatio = wrap.clientWidth / img.naturalWidth;
    const imgHeight = oneImageHeightFromThumbnail * widthRatio;
    img.style.height = imgHeight+"px";
    console.log("Loaded", img.naturalHeight, wrap.clientHeight)
    console.log("Loaded", img.naturalWidth, wrap.clientWidth)
    console.log("Loaded", imgHeight)
}