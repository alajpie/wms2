const tap = require("tap");
const format = require("../utils/format");

tap.equal(format.testables.date(1561556049000), "26.06.2019 15:34");
tap.equal(format.testables.date(1560166643000), "10.06.2019 13:37");
tap.equal(format.testables.date(1546297200000), "1.01.2019 00:00");

tap.equal(format.testables.diff(1560950261138, 1560949971149), "0h 04m 49s");
tap.equal(format.testables.diff(1 + 5 * 60 * 60 * 1000, 1), "5h 00m 00s");
tap.equal(format.testables.diff(1 + 500 * 60 * 60 * 1000, 1), "500h 00m");

const input = [
  { time: 1560954343558, type: "CLOCK_IN" },
  { time: 1561552340579, type: "CLOCK_OUT" },
  { time: 1561553645552, type: "CLOCK_IN" },
  { time: 1561553646614, type: "CLOCK_OUT" },
  { time: 1561557144846, type: "CLOCK_IN" }
];

const currentTime = +new Date();

const output = [
  {
    in: "26.06.2019 15:52",
    out: null,
    duration: format.testables.diff(currentTime, 1561557144846)
  },
  { in: "26.06.2019 14:54", out: "26.06.2019 14:54", duration: "0h 00m 01s" },
  { in: "19.06.2019 16:25", out: "26.06.2019 14:32", duration: "166h 06m" }
];

tap.same(format.entryList(input), output);
