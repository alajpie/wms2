const m = require("mithril");
const css = require("aphrodite").css;
const StyleSheet = require("aphrodite").StyleSheet;
const format = require("../utils/format");

const session = require("../binders/session");
const entries = require("../binders/entries");

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
  entries.refreshStatus();
  list = await entries.list();
  m.redraw();
};

const getStatusState = () => {
  const status = entries.getStatus();
  return status ? status.state : null;
};

const statusClock = () =>
  m(".panel", { class: css(style.panel) }, [
    m(
      "div",
      { class: css(style.status, style.textAlignCenter) },
      entries.getStatus()
        ? [
            m("span", "You're currently "),
            m("b", getStatusState() == "I" ? "clocked in" : "clocked out"),
            m("span", ".")
          ]
        : "Loading..."
    ),
    m("div", { class: css(style.flexCenter, style.clock) }, [
      m(
        "button.btn",
        {
          disabled: getStatusState() != "O",
          class: getStatusState() == "O" ? "btn-primary" : "",
          onclick: e => entries.clockIn().then(refresh)
        },
        "Clock in"
      ),
      m(
        "button.btn",
        {
          disabled: getStatusState() != "I",
          class: getStatusState() == "I" ? "btn-primary" : "",
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
      refresh();
    }, 55 * 1000);
  }
};
