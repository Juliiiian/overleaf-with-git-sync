// Placeholder for future Overleaf toolbar button injection.
// The MutationObserver scaffold below is ready to be extended.

(function () {
  if (document.querySelector('[data-gitsync-injected]')) return;

  const observer = new MutationObserver(() => {
    // TODO: locate the Overleaf toolbar and inject sync buttons.
    // Example selector (inspect your Overleaf version to confirm):
    //   const toolbar = document.querySelector('.toolbar-right');
    // When found: stop observing, inject buttons, mark sentinel.
  });

  observer.observe(document.body, { childList: true, subtree: true });
})();
