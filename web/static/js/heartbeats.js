var clipboard = new ClipboardJS('button.copy-link');
$("button.copy-link").mouseout(function(e) {
    setTimeout(function() {
        e.target.textContent = "copy";
    }, 300);
});
clipboard.on('success', function(e) {
    e.trigger.textContent = "copied!";
    e.clearSelection();
});
clipboard.on('error', function(e) {
    var text = e.trigger.getAttribute("data-clipboard-text");
    prompt("Press Ctrl+C to select:", text)
});
