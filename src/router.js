const m = require("mithril");

const session = require("./models/session");
const Login = require("./routes/login");
const Dash = require("./routes/dash");

module.exports.route = function route() {
  const loggedIn = !!session.token;
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
