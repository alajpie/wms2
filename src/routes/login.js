const m = require("mithril");

const session = require("../models/session");

module.exports = function() {
  let error;

  return {
    view(vnode) {
      return m("div.container", [
        m("a.signin-pb-logo", {
          href: "https://www.bluecrestinc.com"
        }),
        m("div.signin-box", [
          m("h1.text-center.signin-heading", [m("b", "WMS"), m("sup", "2")]),
          m(
            "form",
            {
              onsubmit(e) {
                e.preventDefault();
                session.logIn().catch(e => {
                  if (e.message === "Unauthorized") {
                    error = "Invalid username and/or password.";
                  } else {
                    error = "Connection error, please try again in a while.";
                  }
                  m.redraw();
                });
              }
            },
            [
              m("div.formGroup", [
                m("label", { for: "email" }, "Email"),
                m("input.form-control", {
                  type: "email",
                  oninput(e) {
                    session.email = e.target.value;
                  },
                  value: session.email
                })
              ]),
              m("div.formGroup", [
                m("label", { for: "email" }, "Password"),
                m("input.form-control", {
                  type: "password",
                  oninput(e) {
                    session.password = e.target.value;
                  },
                  value: session.password
                })
              ]),
              m("button.btn.btn-primary", "Submit"),
              m("div", error)
            ]
          )
        ])
      ]);
    }
  };
};
