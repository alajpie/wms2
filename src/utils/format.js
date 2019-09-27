const chunk = require("lodash.chunk");

const z = x => x.toString().padStart(2, 0);

function duration(x) {
  const hours = Math.floor(x / (60 * 60 * 1000));
  const minutes = z(Math.floor((x / (60 * 1000)) % 60));
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

function ms(x) {
  return x * 1000;
}

function entryList(l) {
  const ll = [];
  for (const [day, ens] of Object.entries(l)) {
    const x = {};
    x.date = date(ms(day));

    x.valid = ens.every(x => x.valid);

    const froms = ens.map(x => x.from);
    const first = Math.min(...froms);
    x.from = time(ms(first));

    if (x.valid) {
      const tos = ens.map(x => x.to);
      const last = Math.max(...tos);
      x.to = time(ms(last));
    } else {
      x.to = "XX:XX";
    }

    const valid = ens.filter(x => x.valid);
    if (valid.length === 0) {
      x.duration = "Xh XXm";
    } else {
      const durations = valid.map(x => ms(x.to) - ms(x.from));
      const sum = durations.reduce((x, y) => x + y, 0);
      x.duration = duration(sum);
    }

    x.entries = [];
    for (const en of ens) {
      const y = {};
      y.from = time(ms(en.from));

      y.valid = en.valid;

      if (y.valid) {
        y.to = time(ms(en.to));
        y.duration = duration(ms(en.to) - ms(en.from));
      } else {
        y.to = "XX:XX";
        y.duration = "Xh XXm";
      }

      x.entries.push(y);
    }

    ll.unshift(x);
  }
  return ll;
}

module.exports = { entryList, testables: { duration, date, time } };
