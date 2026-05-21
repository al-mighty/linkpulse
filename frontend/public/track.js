// Cheslav events tracker — sends custom events to LinkPulse /api/events.
//
//   <script async src="https://cheslav.space/linkpulse/track.js"
//           data-project="clm"
//           data-auto-pageview="true"></script>
//
//   window.cheslav.track('button_click', { id: 'demo' });
//
// Drops silently on network errors so the app keeps working.
(function () {
  'use strict';

  var script = document.currentScript || (function () {
    var s = document.getElementsByTagName('script');
    return s[s.length - 1];
  })();
  var ds = (script && script.dataset) || {};

  var PROJECT = ds.project || 'unknown';
  var API = ds.api || 'https://cheslav.space/linkpulse/api/events';
  var AUTO_PAGEVIEW = ds.autoPageview !== 'false';

  function send(name, payload) {
    if (!name) return;
    var body = JSON.stringify({
      project: PROJECT,
      name: String(name).slice(0, 128),
      page: location.pathname + location.search + location.hash,
      payload: payload || null,
    });
    // Prefer sendBeacon for fire-and-forget; fallback to fetch.
    try {
      if (navigator.sendBeacon) {
        var blob = new Blob([body], { type: 'application/json' });
        if (navigator.sendBeacon(API, blob)) return;
      }
    } catch (e) { /* ignore, fall back to fetch */ }
    try {
      fetch(API, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: body,
        keepalive: true,
        mode: 'cors',
        credentials: 'omit',
      }).catch(function () {});
    } catch (e) { /* swallow */ }
  }

  window.cheslav = window.cheslav || {};
  window.cheslav.track = send;
  window.cheslav.project = PROJECT;

  if (AUTO_PAGEVIEW) send('pageview');

  // SPA: re-emit pageview on history navigation.
  var lastPath = location.pathname + location.search + location.hash;
  function maybePageview() {
    var now = location.pathname + location.search + location.hash;
    if (now !== lastPath) {
      lastPath = now;
      if (AUTO_PAGEVIEW) send('pageview');
    }
  }
  window.addEventListener('popstate', maybePageview);
  // Hash routing
  window.addEventListener('hashchange', maybePageview);
  // pushState wrapping for client routers
  var origPush = history.pushState;
  history.pushState = function () {
    origPush.apply(this, arguments);
    setTimeout(maybePageview, 0);
  };
})();
