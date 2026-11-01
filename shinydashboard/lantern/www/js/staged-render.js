/*
  Acknowledges to the server that the filter bar (output "show_filters") has been rendered and its
  inputs registered, so the server can start loading tab data (stage 3 of the staged tab switch in
  server.R) without computing with the previous tab's filter values.

  Protocol:
    1. Server sends the custom message "filters_expected" with a token just before it renders the
       filter bar for a tab that has one.
    2. When "show_filters" arrives and is rendered, wait two animation frames (so the DOM insert and
       Shiny's input binding have run) and send the token back as input$filters_rendered.
  Inputs and the acknowledgement travel over the same ordered websocket, and the acknowledgement is
  a normal (queued) input, so the server sees the new filter values before it sees the token.
  The server has a fallback timer in case this never fires.
*/
(function () {
  var expectedToken = null;

  function onReady(fn) {
    if (window.jQuery && window.Shiny) {
      fn();
    } else {
      document.addEventListener("DOMContentLoaded", fn);
    }
  }

  onReady(function () {
    Shiny.addCustomMessageHandler("filters_expected", function (token) {
      expectedToken = token;
    });

    $(document).on("shiny:value", function (event) {
      if (event.name !== "show_filters" || expectedToken === null) {
        return;
      }
      var token = expectedToken;
      expectedToken = null;
      // shiny:value fires before the content is inserted; two frames lets the render and the
      // input binding finish before the acknowledgement is queued.
      requestAnimationFrame(function () {
        requestAnimationFrame(function () {
          Shiny.setInputValue("filters_rendered", token);
        });
      });
    });
  });
})();
