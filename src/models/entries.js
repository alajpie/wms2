const m = require("mithril");

const session = require("./session");
const consts = require("../consts");

const cache = Symbol("cache");

module.exports = {
  async list() {
    return m.request({
      method: "GET",
      url: consts.API_BASE_URL + "/entries",
      headers: { Authorization: session.token }
    });
  },
  [cache]: {},
  get status() {
    return this[cache].status;
  },
  async refreshStatus() {
    const x = await m.request({
      method: "GET",
      url: consts.API_BASE_URL + "/status",
      headers: { Authorization: session.token }
    });
    this[cache].status = x.status;
    return x.status;
  },
  async clockIn() {
    await m.request({
      method: "PUT",
      url: consts.API_BASE_URL + "/clockin",
      headers: { Authorization: session.token }
    });
    this[cache].status = "CLOCKED_IN";
    m.redraw();
    return;
  },
  async clockOut() {
    await m.request({
      method: "PUT",
      url: consts.API_BASE_URL + "/clockout",
      headers: { Authorization: session.token }
    });
    this[cache].status = "CLOCKED_OUT";
    m.redraw();
    return;
  }
};

window.entries = module.exports; // TODO: remove
