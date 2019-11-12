const m = require("mithril");
const css = require("aphrodite").css;
const StyleSheet = require("aphrodite").StyleSheet;

const session = require("../binders/session");

const style = StyleSheet.create({
  wms: {
    color: "#141b4d",
    fontSize: "300%",
    textAlign: "center"
  },
  separator: {
    height: "20px"
  }
});

module.exports = function() {
  let error;

  return {
    view(vnode) {
      return m("div.container", [
        m("a.signin-pb-logo", {
          href: "https://www.bluecrestinc.com"
        }),
        m(".signin-box", [
          m("div", { class: css(style.wms) }, [m("b", "WMS"), m("sup", "2")]),
          m(
            "form",
            {
              onsubmit(e) {
                e.preventDefault();
                session.logIn().catch(e => {
                  if (e.code === 401) {
                    error = "Invalid username and/or password.";
                  } else {
                    error = "Connection error, check your network connection.";
                  }
                  m.redraw();
                });
              }
            },
            [
              m(".formGroup", [
                m("label", { for: "email" }, "Email"),
                m("input.form-control", {
                  type: "email",
                  oninput(e) {
                    session.setEmail(e.target.value);
                  },
                  value: session.getEmail()
                })
              ]),
              m("div", { class: css(style.separator) }),
              m(".formGroup", [
                m("label", { for: "email" }, "Password"),
                m("input.form-control", {
                  type: "password",
                  oninput(e) {
                    session.password = e.target.value;
                  },
                  value: session.password
                })
              ]),
              m("div", { class: css(style.separator) }),
              m("button.btn.btn-primary", "Submit"),
              m("div", { class: css(error && style.separator) }),
              m("div", error)
            ]
          )
        ])
      ]);
    }
  };
};
