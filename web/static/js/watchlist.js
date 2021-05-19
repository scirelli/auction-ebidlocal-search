export default class Watchlist {
    constructor(rootElem) {
        this.userELem = rootElem.querySelector("#user");
        this.formElem = rootElem.querySelector("#createWatchlistForm");
        this.watchlistElem = rootElem.querySelector("#watchlistList");
        this.addKeywordElem = rootElem.querySelector("#addKeyword");
        this.keywordElem = rootElem.querySelector('#keyword');
        this.submitElem = rootElem.querySelector('input[type="submit"]');
        this.init();
    }

    init() {
        this.addKeywordElem.addEventListener('click', (evt)=> {
            evt.preventDefault();
            let keyword = this.keywordElem.value;
            if(!this.isValidKeyword(keyword)) {
                return this.displayInvalidKeywordError(keyword);
            }

            this.addKeywordElem.disabled = true;
            this.keywordElem.disabled = true;

            let li = document.createElement('li');
            li.innerText = keyword;
            this.watchlistElem.appendChild(li);

            this.addKeywordElem.disabled = false;
            this.keywordElem.disabled = false;
            this.keywordElem.value = '';
            this.keywordElem.focus();
        });

        this.keywordElem.addEventListener("keyup", (evt)=> {
            if(evt.keyCode === 13) {
                evt.preventDefault();
                this.addKeywordElem.click();
            }
        });

        this.formElem.addEventListener('submit', (evt)=> {
            evt.preventDefault();
            evt.stopPropagation();
            return false;
        });
        this.formElem.addEventListener('keypress', (evt)=> {
            if(evt.keyCode === 13) {
                evt.preventDefault();
                evt.stopPropagation();
            }
            return false;
        });

        this.submitElem.addEventListener('click', (evt)=> {
            evt.preventDefault();
            let watchlist = Array.prototype.map.call(this.watchlistElem.querySelectorAll('li'), (li)=> {
                return li.innerText.trim();
            });

            console.log(watchlist);
        });
    }

    isValidKeyword(keyword) {
        if(!keyword) return false;
        let reg = /\W/g

        return !reg.test(keyword)
    }

    displayInvalidKeywordError(keyword) {
        console.error(keyword);
    }
}
