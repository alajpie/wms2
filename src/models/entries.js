const m = require("mithril");

const session = require("./session");
const consts = require("../consts");

const cache = Symbol("cache");

function req(url, method) {
  if (!method) {
    method = "GET";
  }
  return m.request({
    method,
    url: consts.API_BASE_URL + url,
    headers: { Authorization: "Bearer " + session.token }
  });
}

module.exports = {
  async list() {
    return req("/entries");
  },
  [cache]: { status: {} },
  get status() {
    if (!this[cache].status.state) {
      this.refreshStatus();
    }
    return this[cache].status;
  },
  async refreshStatus() {
    const x = await req("/status");
    this[cache].status = x;
    return x;
  },
  async clockIn() {
    await req("/clock/in", "PUT");
    this[cache].status = "CLOCKED_IN";
    m.redraw();
    return;
  },
  async clockOut() {
    await req("/clock/out", "PUT");
    this[cache].status = "CLOCKED_OUT";
    m.redraw();
    return;
  }
};

window.entries = module.exports; // TODO: remove
