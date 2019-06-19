const m = require("mithril");
const css = require("aphrodite").css;
const StyleSheet = require("aphrodite").StyleSheet;
const moment = require("moment");
const chunk = require("lodash.chunk");

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
    fontSize: "3vw"
  },
  gray: {
    color: "#c1c1c1"
  },
  invisible: {
    visibility: "hidden"
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

function formatList(l) {
  const ll = chunk(l, 2)
    .map(x => {
      const z = x => x.toString().padStart(2, 0);
      const f = x => moment(x.time).format("D.MM.YYYY HH:mm");
      if (x[1]) {
        // clocked out
        const d = moment.duration(moment(x[1].time).diff(x[0].time));
        return {
          in: f(x[0]),
          out: f(x[1]),
          duration: Math.floor(d.asHours()) + "h " + z(d.minutes()) + "m"
        };
      } else {
        // clocked in
        const d = moment.duration(moment().diff(x[0].time));
        return {
          in: f(x[0]),
          out: null,
          duration: Math.floor(d.asHours()) + "h " + z(d.minutes()) + "m"
        };
      }
    })
    .reverse();
  return ll;
}

const listView = list =>
  m(
    ".panel",
    {
      class: css(style.panel, style.textAlignCenter)
    },
    formatList(list).map(x => {
      if (x.out) {
        return m(
          "div",
          { class: css(style.entry) },
          `${x.in}–${x.out} (${x.duration})`
        );
      } else {
        return m("div", { class: css(style.entry) }, [
          m("span", `${x.in}–`),
          m("span", { class: css(style.invisible) }, "12.34.5678 90:12"),
          m("span", { class: css(style.gray) }, ` (${x.duration})`)
        ]);
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
