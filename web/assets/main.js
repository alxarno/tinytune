const imageOnload = (id) => {
    const img = document.getElementById(`video-${id}`);
    const duration = document.getElementById(`duration-${id}`)
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
    duration.style.right = (wrap.clientWidth - imgWidth) / 2 + "px"
}

const onZoomChanged = () => {
    const items = Array.from(document.getElementsByClassName("preview"))
    items.forEach(v => imageOnload(v.id.replace("video-", "")))
}

const zoom = {
    state: Cookies.get('zoom') || 'medium',
    transitions: {
        medium: {
            in() {
                this.state = 'large'
            },
            out() {
                this.state = 'small'
            }
        },
        small: {
            in() {
                this.state = 'medium'
            },
            out() {
                this.state = 'xs'
            }
        },
        large: {
            in() {
                this.state = 'xl'
            },
            out() {
                this.state = 'medium'
            }
        },
        xl: {
            out() {
                this.state = 'large'
            },
        },
        xs: {
            in() {
                this.state = 'small'
            }
        }
    },
    dispatch(actionName) {
        const oldState = this.state
        const action = this.transitions[this.state][actionName];
        if (action) {
            action.call(this);
            Cookies.set('zoom', this.state);
            document.body.classList.replace(`zoom-${oldState}`, `zoom-${this.state}`)
            onZoomChanged()
        }
    },
};

const onSort = (radio) => {
    Cookies.set('sort', radio.value);
    htmx.ajax('GET', window.location.pathname)
}

const onSearch = () => {
    let url = window.location.pathname
    const searchInput = document.getElementById("search-input")
    url = url.replace("/d/", "/s/")
    if (!url.includes("/s")) {
        url += "s"
    }
    url += `?query=${encodeURIComponent(searchInput.value)}`
    htmx.ajax('GET', url).then((event) => {
        highlightSearchResults()
    })
}

const highlightSearchResults = () => {
    const foundElement = document.getElementById("found")
    const searchInput = document.getElementById("search-input")
    if(!foundElement) return;
    const labels =  Array.from(document.getElementsByClassName("figure-caption"))
    labels.forEach((element) => {
        const start = element.textContent.indexOf(searchInput.value)
        const htmlValue = `${element.textContent.substring(0, start)}<span class="bg-primary text-dark rounded-1">${element.textContent.substring(start, start + searchInput.value.length)}</span>${element.textContent.substring(start + searchInput.value.length)}`
        element.innerHTML = htmlValue
    })
}

window.onload = () => {
    highlightSearchResults()
}