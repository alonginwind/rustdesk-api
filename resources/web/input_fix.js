// Work around the Flutter web engine mis-reporting an IME composition inside the
// terminal's hidden text editing element.
//
// The engine derives the composition start as `selectionEnd - preedit.length`
// (text_editing/composition_aware_mixin.dart). In that element the selection sits
// at 0 while Chrome inserts the preedit, so the start is negative, the engine
// reports a collapsed composition, and xterm treats the pinyin as ordinary text:
// it types the letters into the PTY and then resets the element, which aborts the
// composition and drops the candidate window.
//
// While the engine would mis-detect, hide the preedit updates from it; when the
// composition ends, hand the committed text over as one ordinary `input` event so
// the engine reads the final value and the terminal receives whole characters.
(function () {
  var target = null;
  var preedit = '';
  var hiding = false;

  // Same rule the engine applies: a negative composition start means "no
  // composition", i.e. the preedit would be taken for plain text.
  function engineSeesCollapsed(el) {
    return !!el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') &&
      el.selectionEnd - preedit.length < 0;
  }

  // Capture on window, so these run before the engine's own handlers on the
  // element and on the document.
  window.addEventListener('compositionstart', function (e) {
    target = e.target;
    preedit = '';
    hiding = false;
  }, true);

  // Latch it: once this composition is known to be mis-detected, keep hiding
  // until it ends even if the element's selection moves in the meantime.
  window.addEventListener('compositionupdate', function (e) {
    if (e.target !== target) return;
    preedit = e.data || '';
    hiding = hiding || engineSeesCollapsed(target);
  }, true);

  window.addEventListener('input', function (e) {
    if (hiding && e.target === target) e.stopImmediatePropagation();
  }, true);

  window.addEventListener('selectionchange', function (e) {
    if (hiding) e.stopImmediatePropagation();
  }, true);

  window.addEventListener('compositionend', function (e) {
    var wasHiding = hiding;
    hiding = false;
    preedit = '';
    if (wasHiding && e.target === target) {
      target.dispatchEvent(new Event('input', { bubbles: true }));
    }
  }, true);
})();
