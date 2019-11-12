const m = require("mithril");

const css = require("aphrodite").css;
const StyleSheet = require("aphrodite").StyleSheet;

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
  panel: {
    "@media (min-width:600px)": {
      padding: "30px"
    },
    padding: "15px",
    maxWidth: "600px",
    margin: "0 auto"
  }
});

module.exports = {
  view(vnode) {
    return m(
      "div",
      { class: css(style.font) },
      m(
        ".panel",
        { class: css(style.panel) },
        m(
          "div",
          { class: css(style.status, style.textAlignCenter) },
          "Something went wrong. Try refreshing?"
        )
      )
    );
  }
};
