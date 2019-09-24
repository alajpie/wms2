const m = require("mithril");

const session = require("./binders/session");
const Login = require("./views/login");
const Dash = require("./views/dash");

module.exports.route = function route() {
  const loggedIn = !!session.getToken();
  if (loggedIn) {
    m.route(document.body, "/dash", {
      "/dash": Dash
    });
  } else {
    m.route(document.body, "/login", {
      "/login": Login
    });
  }
};
