const m = require("mithril");
const css = require("aphrodite").css;
const StyleSheet = require("aphrodite").StyleSheet;
const format = require("../utils/format");

const session = require("../binders/session");
const entries = require("../binders/entries");

const style = StyleSheet.create({
  font: {
    "@media (min-width:600px)": {
      fontSize: "15px"
    },
    fontSize: "2.5vw"
  },
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
    color: "#000000",
    fontSize: "200%"
  },
  clock: {
    marginTop: "15px"
  },
  panel: {
    "@media (min-width:600px)": {
      padding: "30px"
    },
    padding: "15px",
    maxWidth: "600px",
    margin: "0 auto"
  },
  separator: {
    height: "20px"
  },
  day: {
    color: "black",
    fontSize: "180%",
    borderRadius: "5px",
    display: "flex",
    justifyContent: "space-evenly"
  },
  entry: {
    width: "70%",
    marginLeft: "15%",
    color: "black",
    fontSize: "180%",
    borderRadius: "5px",
    display: "flex",
    justifyContent: "space-evenly"
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

const statusClock = () => {
  const status = entries.getStatus();
  const statusState = status ? status.state : null;
  return m(".panel", { class: css(style.panel) }, [
    status
      ? m("div", [
          m("div", { class: css(style.textAlignCenter) }, [
            m("b", status.online),
            status.online === 1
              ? m("span", " person is clocked in right now.")
              : m("span", " people are clocked in right now.")
          ]),
          m(
            "div",
            { class: css(style.status, style.textAlignCenter) },
            statusState === "I"
              ? [
                  m("span", "You've been "),
                  m("b", "clocked in"),
                  m("span", " for "),
                  m("b", format.duration(Date.now() - status.since * 1000)),
                  m("span", ".")
                ]
              : [
                  m("span", "You're currently "),
                  m("b", "clocked out"),
                  m("span", ".")
                ]
          ),
          m("div", { class: css(style.status, style.textAlignCenter) }, [
            m("span", "You're "),
            m("b", format.duration((status.deltaForDay - status.since) * 1000)),
            m(
              "b",
              status.deltaForDay - (Date.now() / 1000 - status.since) < 0
                ? " behind"
                : " ahead"
            ),
            m("span", " for the day.")
          ]),
          m("div", { class: css(style.status, style.textAlignCenter) }, [
            m("span", "You're "),
            m(
              "b",
              format.duration((status.deltaForMonth - status.since) * 1000)
            ),
            m(
              "b",
              status.deltaForMonth - (Date.now() / 1000 - status.since) < 0
                ? " behind"
                : " ahead"
            ),
            m("span", " for the month.")
          ])
        ])
      : m("div", [
          m("div", { class: css(style.textAlignCenter) }, [
            m("b", "[...]"),
            m("span", " people are clocked in right now")
          ]),
          m("div", { class: css(style.status, style.textAlignCenter) }, [
            m("span", "The page is "),
            m("b", "loading"),
            m("span", ".")
          ]),
          m("div", { class: css(style.status, style.textAlignCenter) }, [
            m("span", "You're "),
            m("b", "[...]"),
            m("span", " for the day.")
          ]),
          m("div", { class: css(style.status, style.textAlignCenter) }, [
            m("span", "You're "),
            m("b", "[...]"),
            m("span", " for the month.")
          ])
        ]),
    m("div", { class: css(style.flexCenter, style.clock) }, [
      m(
        "button.btn",
        {
          disabled: statusState !== "O",
          class: statusState === "O" ? "btn-primary" : "",
          onclick: e => entries.clockIn().then(refresh)
        },
        "Clock in"
      ),
      m(
        "button.btn",
        {
          disabled: statusState !== "I",
          class: statusState === "I" ? "btn-primary" : "",
          onclick: e => entries.clockOut().then(refresh)
        },
        "Clock out"
      )
    ])
  ]);
};

const day = day =>
  m("div", [
    m(
      "div",
      {
        class: css(style.day, style.shaded, !day.valid && style.red)
      },
      [
        m("span", day.date),
        m("span", day.duration),
        m("span", `(${day.from} - ${day.to})`)
      ]
    ),
    day.entries.map(x => entry(x))
  ]);

const entry = entry =>
  m(
    "div",
    {
      class: css(style.entry, !entry.valid && style.red)
    },
    [m("span", `${entry.from} - ${entry.to}`), m("span", `(${entry.duration})`)]
  );

const listView = list =>
  m(
    ".panel",
    {
      class: css(style.panel, style.textAlignCenter)
    },
    format.entryList(list).map(x => day(x))
  );

module.exports = {
  view(vnode) {
    return m("div", { class: css(style.font) }, [
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
