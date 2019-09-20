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
    maxWidth: "600px",
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
  red: {
    color: "#ff0000"
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
      entries.status.state
        ? [
            m("span", "You're currently "),
            m("b", entries.status.state == "I" ? "clocked in" : "clocked out"),
            m("span", ".")
          ]
        : "Loading..."
    ),
    m("div", { class: css(style.flexCenter, style.clock) }, [
      m(
        "button.btn",
        {
          disabled: entries.status.state != "O",
          class: entries.status.state == "O" ? "btn-primary" : "",
          onclick: e => entries.clockIn().then(refresh)
        },
        "Clock in"
      ),
      m(
        "button.btn",
        {
          disabled: entries.status.state != "I",
          class: entries.status.state == "I" ? "btn-primary" : "",
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
    format.entryList(list).map((x, i) =>
      x.entries.map((y, j) =>
        m(
          "div",
          {
            class: css(
              style.entry,
              j % 2 == 0 && style.shaded,
              !y.valid && style.red
            )
          },
          `${y.date} ${y.duration} (${y.from} – ${y.to})`
        )
      )
    )
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
