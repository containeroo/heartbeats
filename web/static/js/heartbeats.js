// copy handler: clicking the cell copies its URL
function attachCopyHandlers() {
  document.querySelectorAll(".copy-cell").forEach((cell) => {
    cell.addEventListener("click", () => {
      const text = cell.querySelector(".url-text").innerText.trim();
      navigator.clipboard
        .writeText(text)
        .then(() => {
          const btn = cell.querySelector(".copy-btn");
          btn.textContent = "Copied!";
          setTimeout(() => (btn.textContent = "Copy"), 1500);
        })
        .catch(console.error);
    });
  });
}

/**
 * Wire up a text‐search + clear "×" inside an .input-group.filter.
 *
 * @param {object} opts
 * @param {string} opts.inputId    – the <input id="…">
 * @param {string} opts.rowSelector – CSS selector matching each row to show/hide
 * @param {(row: Element)=>string} opts.getText  – given a row, return the text to match
 */
function attachFilter({ inputId, rowSelector, getText }) {
  const input = document.getElementById(inputId);
  if (!input) return;
  const wrapper = input.closest(".input-group.filter");
  if (!wrapper) return;
  const clearBtn = wrapper.querySelector(".clear-filter-btn");
  if (!clearBtn) return;

  const filterRows = () => {
    const needle = input.value.trim().toLowerCase();
    document.querySelectorAll(rowSelector).forEach((row) => {
      const txt = getText(row).trim().toLowerCase();
      row.style.display = !needle || txt.includes(needle) ? "" : "none";
    });
  };

  // show/hide clear-"×" and re-filter on every keystroke
  input.addEventListener("input", () => {
    clearBtn.style.display = input.value ? "block" : "none";
    filterRows();
  });

  // clear everything when "×" is clicked
  clearBtn.addEventListener("click", () => {
    input.value = "";
    clearBtn.style.display = "none";
    input.focus();
    // reset all rows visible
    document.querySelectorAll(rowSelector).forEach((row) => {
      row.style.display = "";
    });
  });

  // initialize
  clearBtn.style.display = input.value ? "block" : "none";
}

/**
 * Fetches and injects the partial, then re-wires all UI bits.
 */
async function loadSection(sec) {
  try {
    const res = await fetch(`/partials/${sec}`);
    const html = await res.text();
    document.getElementById("content").innerHTML = html;

    // Nav highlight
    document
      .querySelectorAll("nav .nav-link")
      .forEach((a) => a.classList.remove("active"));
    document
      .querySelector(`nav .nav-link[data-section="${sec}"]`)
      ?.classList.add("active");

    // Cleanup old tooltips & init new ones
    document.querySelectorAll(".tooltip").forEach((t) => t.remove());
    document.querySelectorAll('[data-bs-toggle="tooltip"]').forEach((el) => {
      const old = bootstrap.Tooltip.getInstance(el);
      if (old) old.dispose();
      new bootstrap.Tooltip(el, {
        trigger: "hover focus",
        delay: { show: 100, hide: 100 },
        container: "body",
      });
    });

    // Copy-to-clipboard
    attachCopyHandlers();

    // Heartbeats table: match by second <td>
    attachFilter({
      inputId: "searchHeartbeat",
      rowSelector: ".table-wrapper tbody tr",
      getText: (row) => row.children[1].textContent,
    });

    // Receivers table: match by data-id
    attachFilter({
      inputId: "searchReceiver",
      rowSelector: "#recv-body tr",
      getText: (row) => row.dataset.id,
    });

    // History table: also match by data-id
    attachFilter({
      inputId: "searchHistory",
      rowSelector: "#hist-body tr",
      getText: (row) => row.dataset.id,
    });

    // Sort the heartbeats table if present
    const table = document.querySelector(".table-wrapper table");
    if (table) {
      new Tablesort(table);
    }
  } catch (err) {
    console.error("loadSection error:", err);
  }
}

/** Jump to the Receivers tab and prefill its search box. */
async function goToReceiver(rid) {
  await loadSection("receivers");
  window.location.hash = "receivers";

  const inpt = document.getElementById("searchReceiver");
  if (!inpt) return;
  inpt.value = rid;
  inpt.dispatchEvent(new Event("input"));
}

/** Jump to the History tab and prefill its search box. */
async function goToHistory(hb) {
  await loadSection("history");
  window.location.hash = "history";

  const inpt = document.getElementById("searchHistory");
  if (!inpt) return;
  inpt.value = hb;
  inpt.dispatchEvent(new Event("input"));
}

document.addEventListener("DOMContentLoaded", () => {
  // Wire the nav‐links:
  document.querySelectorAll("nav .nav-link").forEach((a) => {
    a.addEventListener("click", (e) => {
      e.preventDefault(); // don’t jump
      const sec = a.dataset.section; // e.g. "receivers"
      if (!sec) return;
      loadSection(sec);
      // update the hash so refresh will know:
      window.location.hash = sec;
    });
  });

  // initial page load, pick up the hash (if any)
  let initial = window.location.hash.slice(1); // strip the "#"
  if (!initial || !["heartbeats", "receivers", "history"].includes(initial)) {
    initial = "heartbeats";
  }
  loadSection(initial);
});
