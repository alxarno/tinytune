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
        const startElement = element.textContent.substring(0, start)
        const highlightedElement = element.textContent.substring(start, start + searchInput.value.length)
        const endElement = element.textContent.substring(start + searchInput.value.length)
        const htmlValue = `${startElement}<span class="bg-primary text-dark rounded-1">${highlightedElement}</span>${endElement}`
        element.innerHTML = htmlValue
    })
}

window.onload = () => {
    highlightSearchResults()
}