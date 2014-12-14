'use strict';

var СhanList = function (data) {
    var lst = [];
    var current = 0;
    var i = 0;
    var loaded = false;
    for (var c in data) {
        if (data.hasOwnProperty(c)) {
            lst[i++] = {
                'title': data[c],
                'url': c
            };
        }
    }
    loaded = true;
    var next = function () {
        if (current === lst.length - 1) {
            current = 0
        } else {
            current++
        }
    };
    var prev = function () {
        if (current > 0) {
            current--
        } else {
            current = lst.length - 1
        }
    };
    var getCurrent = function () {
        return lst[current]
    };
    var setCurrent = function (newCur) {
        current = newCur
    };
    var render = function (container, callback) {
        var list = document.getElementById(container);
        list.innerHTML = '';
        var i = 0;
        for (var c in lst) {
            if (lst.hasOwnProperty(c)) {
                var li = document.createElement('li');
                var title = document.createTextNode(lst[c].title);
                if (current == i) {
                    li.className = 'current';
                }
                li.appendChild(title);
                li.setAttribute('data-id', i++);
                li.setAttribute('data-url', lst[c].url);
                li.onclick = callback;
                list.appendChild(li);
            }
        }
    };
    return {
        list: lst,
        current: current,
        next: next,
        prev: prev,
        getCurrent: getCurrent,
        setCurrent: setCurrent,
        render: render,
        loaded: loaded
    };
};

function onHLSReady() {
    var url = BASE_URL + ccc.getCurrent().url + SUFFIX;
    player = getPlayer(player_id);
    if (player != null) {
        player.height = parseInt(window.innerHeight);
        load(url);
        play();
    }
}

function chanSwitch() {
    player.playerStop();
    var id = parseInt(this.getAttribute('data-id'));
    ccc.setCurrent(id);
    load(BASE_URL + ccc.getCurrent().url + SUFFIX);
    play();
    ccc.render('list', chanSwitch);
}

function getChannels(url, success, error) {
    var ajax = new XMLHttpRequest();
    ajax.open('GET', url, true);
    ajax.responseType = 'json';
    ajax.onreadystatechange = function () {
        if (ajax.readyState == 4) {
            if (ajax.status == 200) {
                success && success(ajax.response);
            } else {
                error && error(ajax.status);
            }
        }
    };
    ajax.send();
}

function getPlayer(id) {
    if (window.document[id]) {
        return window.document[id];
    }
    if (navigator.appName.indexOf("Microsoft Internet") == -1) {
        if (document.embeds && document.embeds[id]) {
            return document.embeds[id];
        }
    } else {
        return document.getElementById(id);
    }
}

var ccc;

getChannels(
    BASE_URL + '/channels.json',
    function (data) {
        ccc = new СhanList(data);
        console.log(ccc.getCurrent().url);
        ccc.render('list', chanSwitch);
    },
    function () {
        var list = document.getElementById('list');
        var li = document.createElement('li');
        var title = document.createTextNode('Ошибка загрузки :(');
        li.appendChild(title);
        list.appendChild(li);
    }
);

function load(url) {
    player.playerStop();
    player.playerLoad(url);
}

function play() {
    player.playerPlay(-1);
}

function onError(code, url, msg) {
    console.dir({code: code, url: url, msg: msg});
}

document.onkeyup = function (evt) {
    if (evt.keyCode === 40) {
        console.log('next channel');
        ccc.next();
        load(BASE_URL + ccc.getCurrent().url + SUFFIX);
        play();
        ccc.render('list', chanSwitch);
    } else if (evt.keyCode === 38) {
        console.log('prev channel');
        ccc.prev();
        load(BASE_URL + ccc.getCurrent().url + SUFFIX);
        play();
        ccc.render('list', chanSwitch);
    }
};
