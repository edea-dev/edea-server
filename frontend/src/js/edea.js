 
function iterateOverHTMLCollectionAndMarkActive(uwu) {
    for (var i = 0; i < uwu.length; i++) {
        let OwO = uwu.item(i);
        if (OwO.tagName.toLowerCase() != "a") {
            continue;
        }
        if (window.location.href.startsWith(OwO.href)) {
            OwO.classList.add("active");
            let srhelper = document.createElement('span');
            srhelper.innerText = "current page is: ";
            srhelper.classList.add("visually-hidden");
            OwO.parentNode.insertBefore(srhelper, OwO);
        }
    }
}

function highlightActiveMenuItem() {
    iterateOverHTMLCollectionAndMarkActive(
        document.getElementsByClassName("dropdown-item"));
    iterateOverHTMLCollectionAndMarkActive(
        document.getElementsByClassName("nav-link"));
}

highlightActiveMenuItem();