window._gwen = {}
window._gwen.kv = {}

if (!localStorage.getItem('option:local:theme')) {
    localStorage.setItem('option:local:theme', 'dark')
}
if (!localStorage.getItem('option:local:lang')) {
    localStorage.setItem('option:local:lang', 'zh-CN')
}

function getQueryVariable() {
    var hash = window.location.hash;
    var qIdx = hash.indexOf('?');
    if (qIdx < 0) return;
    var query = hash.substring(qIdx + 1);
    var vars = query.split("&");
    for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split("=");
        window._gwen.kv[pair[0]] = pair[1]
    }
}
getQueryVariable()

const share_token = window._gwen.kv.share_token || ''
if (share_token) {
    fetch("/api/shared-peer", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({share_token})
    }).then(res => res.json()).then(res => {
        if (res.code === 0) {
            localStorage.setItem('option:local:share_source', 'share')
            const peer = res.data.peer || {}
            window.location.href = `/webclient/#/${peer.info.id}?password=${encodeURIComponent(peer.tmppwd)}`
        }
    })
}
