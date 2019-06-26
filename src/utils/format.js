const chunk = require("lodash.chunk");

const z = x => x.toString().padStart(2, 0);

function diff(x, y) {
  const diff = new Date(+x).valueOf() - new Date(+y).valueOf();
  const hours = Math.floor(diff / (60 * 60 * 1000));
  const minutes = z(Math.floor((diff / (60 * 1000)) % 60));
  const seconds = z(Math.floor((diff / 1000) % 60));
  if (hours < 100) {
    return `${hours}h ${minutes}m ${seconds}s`;
  } else {
    return `${hours}h ${minutes}m`;
  }
}

function date(x) {
  const t = new Date(x);
  return (
    t.getDate() +
    "." +
    z(t.getMonth() + 1) +
    "." +
    t.getFullYear() +
    " " +
    z(t.getHours()) +
    ":" +
    z(t.getMinutes())
  );
}

function entryList(l) {
  const ll = chunk(l, 2)
    .map(x => {
      return {
        in: date(x[0].time),
        out: x[1] ? date(x[1].time) : null,
        duration: x[1]
          ? diff(x[1].time, x[0].time)
          : diff(Date.now(), x[0].time)
      };
    })
    .reverse();
  return ll;
}

module.exports = { entryList, testables: { diff, date } };
