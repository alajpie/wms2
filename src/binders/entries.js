const m = require("mithril");

const session = require("./session");
const consts = require("../consts");

function req(url, method) {
  if (!method) {
    method = "GET";
  }
  return m.request({
    method,
    url: consts.API_BASE_URL + url,
    headers: { Authorization: "Bearer " + session.getToken() }
  });
}

let cachedStatus = null;
let cachedExpiry = 0;

module.exports = {
  async list() {
    return req("/entries");
  },
  getStatus() {
    if (cachedExpiry < Date.now()) {
      return null;
    } else {
      return cachedStatus;
    }
  },
  async refreshStatus() {
    cachedExpiry = Date.now() + 60 * 1000;
    cachedStatus = await req("/status");
    return cachedStatus;
  },
  async clockIn() {
    await req("/clock/in", "PUT");
    cachedExpiry = 0;
    this.refreshStatus();
    m.redraw();
    return;
  },
  async clockOut() {
    await req("/clock/out", "PUT");
    cachedExpiry = 0;
    this.refreshStatus();
    m.redraw();
    return;
  }
};

window.entries = module.exports; // TODO: remove
