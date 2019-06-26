const m = require("mithril");
const css = require("aphrodite").css;
const StyleSheet = require("aphrodite").StyleSheet;
const format = require("../utils/format");

const session = require("../models/session");
const entries = require("../models/entries");

const style = StyleSheet.create({
  flexRight: {
    width: "100%",
    display: "flex",
    justifyContent: "flex-end"
  },
  flexCenter: {
    display: "flex",
    justifyContent: "center"
  },
  textAlignCenter: {
    textAlign: "center"
  },
  status: {
    color: "black",
    fontSize: "200%"
  },
  clock: {
    marginTop: "15px"
  },
  panel: {
    padding: "30px",
    width: "80%",
    maxWidth: "800px",
    margin: "0 auto"
  },
  separator: {
    height: "20px"
  },
  entry: {
    color: "black",
    "@media (min-width:1000px)": {
      fontSize: "30px"
    },
    fontSize: "3vw",
    borderRadius: "5px"
  },
  ghost: {
    color: "#adadad"
  },
  invisible: {
    visibility: "hidden"
  },
  shaded: {
    backgroundColor: "#dedede"
  }
});

let list = [];

const refresh = async () => {
  const x = entries.refreshStatus();
  const y = entries.list();
  await x;
  list = await y;
  m.redraw();
};

const statusClock = () =>
  m(".panel", { class: css(style.panel) }, [
    m(
      "div",
      { class: css(style.status, style.textAlignCenter) },
      entries.status
        ? [
            m("span", "You're currently "),
            m(
              "b",
              entries.status == "CLOCKED_IN" ? "clocked in" : "clocked out"
            ),
            m("span", ".")
          ]
        : "Loading..."
    ),
    m("div", { class: css(style.flexCenter, style.clock) }, [
      m(
        "button.btn",
        {
          disabled: entries.status != "CLOCKED_OUT",
          class: entries.status == "CLOCKED_OUT" ? "btn-primary" : "",
          onclick: e => entries.clockIn().then(refresh)
        },
        "Clock in"
      ),
      m(
        "button.btn",
        {
          disabled: entries.status != "CLOCKED_IN",
          class: entries.status == "CLOCKED_IN" ? "btn-primary" : "",
          onclick: e => entries.clockOut().then(refresh)
        },
        "Clock out"
      )
    ])
  ]);

const listView = list =>
  m(
    ".panel",
    {
      class: css(style.panel, style.textAlignCenter)
    },
    format.entryList(list).map((x, i) => {
      if (x.out) {
        return m(
          "div",
          { class: css(style.entry, i % 2 == 0 && style.shaded) },
          `${x.in} – ${x.out} (${x.duration})`
        );
      } else {
        return m(
          "div",
          { class: css(style.entry, i % 2 == 0 && style.shaded) },
          [
            m("span", `${x.in} – `),
            m("span", { class: css(style.invisible) }, "12.34.5678 90:12"),
            m("span", { class: css(style.ghost) }, ` (${x.duration})`)
          ]
        );
      }
    })
  );

module.exports = {
  view(vnode) {
    return m("div", [
      m(
        "#header",
        m(".navbar", [
          m(
            "span",
            { class: css(style.flexRight) },
            m("button.btn", { onclick: e => session.logOut() }, "Log out")
          )
        ])
      ),
      statusClock(),
      m("div", { class: css(style.separator) }),
      listView(list)
    ]);
  },
  async oninit() {
    refresh();
    setInterval(() => {
      entries.status === "CLOCKED_IN" && m.redraw();
    }, 1000);
  }
};
