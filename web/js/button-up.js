let button = null

const onButtonUpClick = () => {
    document.body.scrollTop = 0;
    document.documentElement.scrollTop = 0;
}

const scrollFunction = () => {
    if (document.body.scrollTop > 20 || document.documentElement.scrollTop > 20) {
        button.style.display = "block";
    } else {
        button.style.display = "none";
    }
  }

const buttonUpInit = () => {
    window.onscroll = function() {scrollFunction()};
    button = document.getElementById("button-up");
}

export {buttonUpInit, onButtonUpClick}