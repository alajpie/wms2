const m = require("mithril");

const session = require("./binders/session");
const Error = require("./views/error");
const Login = require("./views/login");
const Dash = require("./views/dash");

module.exports.route = function route() {
  session.versionMatches().then(ok => {
    if (!ok) {
      m.route(document.body, "/error", { "/error": Error });
    }
  });
  const loggedIn = session.getToken() !== "null";
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
