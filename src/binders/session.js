const m = require("mithril");

const router = require("../router");
const consts = require("../consts");

module.exports = {
  getEmail() {
    return window.localStorage.getItem("email");
  },
  setEmail(x) {
    window.localStorage.setItem("email", x);
  },
  getToken() {
    return window.localStorage.getItem("token");
  },
  setToken(x) {
    window.localStorage.setItem("token", x);
  },

  password: undefined,

  async logIn() {
    const x = await m.request({
      method: "POST",
      url: consts.API_BASE_URL + "/authorize",
      body: { email: this.getEmail(), password: this.password }
    });
    this.setToken(x.token);
    this.password = null;
    router.route();
  },
  logOut() {
    this.setToken(null);
    this.password = null;
    router.route();
  },

  async onlineUsers() {
    return await m.request({
      method: "GET",
      url: consts.API_BASE_URL + "/u/users/online/count",
      headers: { Authorization: "Bearer " + this.getToken() }
    });
  }
};

window.session = module.exports; // TODO: remove
