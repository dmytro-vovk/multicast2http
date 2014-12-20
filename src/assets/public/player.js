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
    var url = channels.getCurrent().url + SUFFIX;
    player = getPlayer(player_id);
    if (player != null) {
        player.height = parseInt(window.innerHeight);
        load(url);
    }
}

function chanSwitch() {
    var id = parseInt(this.getAttribute('data-id'));
    channels.setCurrent(id);
    load(channels.getCurrent().url + SUFFIX);
    channels.render('list', chanSwitch);
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

var channels;

getChannels(
    '/channels.json',
    function (data) {
        channels = new СhanList(data);
        channels.render('list', chanSwitch);
        if (useHtml5) {
            load(channels.getCurrent().url + SUFFIX);
        }
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
    if (useHtml5) {
        document.getElementById('player-container').innerHTML =
            '<video id="vplayer" width="100%" height="100%" src="' + url + '" autoplay type="application/x-mpegURL"></video>';
        var vplayer = document.getElementById('vplayer');
        vplayer.play();
        vplayer.load();
        vplayer.play();
    } else {
        player.playerStop();
        player.playerLoad(url);
        player.playerPlay(-1);
    }
}

function onError(code, url, msg) {
    console.dir({code: code, url: url, msg: msg});
}

document.onkeyup = function (evt) {
    if (evt.keyCode === 40) {
        console.log('next channel');
        channels.next();
        load(channels.getCurrent().url + SUFFIX);
        channels.render('list', chanSwitch);
    } else if (evt.keyCode === 38) {
        console.log('prev channel');
        channels.prev();
        load(channels.getCurrent().url + SUFFIX);
        channels.render('list', chanSwitch);
    }
};

window.onresize = function () {
    var w = parseInt(window.innerWidth), h = parseInt(window.innerHeight);
    if (useHtml5) {
        var vplayerContainer = document.getElementById('player-container');
        var vplayer = document.getElementById('vplayer');
        vplayerContainer.style.width = w + 'px';
        vplayerContainer.style.height = h + 'px';
        vplayer.style.width = w + 'px';
        vplayer.style.height = h + 'px';
    } else {
        player.height = h;
        player.width = w;
    }
    document.getElementById('list').style.maxHeight = (parseInt(window.innerHeight) - 20 - 20 - 20 - 7 - 7) + 'px';
};
