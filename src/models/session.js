const m = require("mithril");

const router = require("../router");
const consts = require("../consts");

const email = Symbol("email");
const token = Symbol("token");

module.exports = {
  [email]: window.localStorage.getItem("email"),
  get email() {
    return this[email];
  },
  set email(x) {
    window.localStorage.setItem("email", x ? x : "");
    this[email] = x;
  },
  password: undefined,
  [token]: window.localStorage.getItem("token"),
  get token() {
    return this[token];
  },
  set token(x) {
    window.localStorage.setItem("token", x ? x : "");
    this[token] = x;
  },

  async logIn() {
    const x = await m.request({
      method: "POST",
      url: consts.API_BASE_URL + "/authorize",
      body: { email: this.email, password: this.password }
    });
    this.token = x.token;
    this.password = null;
    router.route();
  },
  logOut() {
    this.token = null;
    this.password = null;
    router.route();
  }
};

window.session = module.exports; // TODO: remove
