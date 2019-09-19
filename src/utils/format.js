const chunk = require("lodash.chunk");

const z = x => x.toString().padStart(2, 0);

function diff(x, y) {
  const diff = new Date(+x).valueOf() - new Date(+y).valueOf();
  const hours = Math.floor(diff / (60 * 60 * 1000));
  const minutes = z(Math.floor((diff / (60 * 1000)) % 60));
  return `${hours}h ${minutes}m`;
}

function date(x) {
  const t = new Date(x);
  return t.getDate() + "." + z(t.getMonth() + 1) + "." + t.getFullYear();
}

function time(x) {
  const t = new Date(x);
  return z(t.getHours()) + ":" + z(t.getMinutes());
}

function entryList(l) {
  const ll = l
    .map(x => {
      return {
        date: date(x.from * 1000),
        from: time(x.from * 1000),
        to: x.valid ? time(x.to * 1000) : "XXXX",
        duration: x.valid ? diff(x.to * 1000, x.from * 1000) : "Xh XXm",
        valid: x.valid
      };
    })
    .reverse();
  return ll;
}

module.exports = { entryList, testables: { diff, date } };
