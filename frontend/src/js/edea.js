 
function highlightActiveMenuItem() {
    const navelems = document.getElementsByClassName("nav-link");
    for (var i = 0; i < navelems.length; i++) {
        let currentNavE = navelems.item(i);
        if (window.location.href.startsWith(currentNavE.href)) {
            currentNavE.classList.add("active");
            let srhelper = document.createElement('span');
            srhelper.innerText = "current page is: ";
            srhelper.classList.add("visually-hidden");
            currentNavE.parentNode.insertBefore(srhelper, currentNavE);
        }
    }
}

highlightActiveMenuItem();