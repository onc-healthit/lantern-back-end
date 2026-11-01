/*
  Tells the server whenever a modal has finished closing (input$modal_closed). server.R uses this
  to reset which endpoint-modal sub-tabs count as loaded (loaded_modal_tabs), so the reset happens
  while no modal is open instead of in the same flush as the next popup.
*/
(function () {
  function onReady(fn) {
    if (window.jQuery && window.Shiny) {
      fn();
    } else {
      document.addEventListener("DOMContentLoaded", fn);
    }
  }

  onReady(function () {
    $(document).on("hidden.bs.modal", ".modal", function () {
      Shiny.setInputValue("modal_closed", Date.now(), { priority: "event" });
    });
  });
})();
