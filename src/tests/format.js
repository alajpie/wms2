const tap = require("tap");
const format = require("../utils/format");

tap.test("date and time", test => {
  test.equal(format.testables.date(1561556049000), "26.06.2019");
  test.equal(format.testables.time(1561556049000), "15:34");
  test.equal(format.testables.date(1560166643000), "10.06.2019");
  test.equal(format.testables.time(1560166643000), "13:37");
  test.equal(format.testables.date(1546297200000), "1.01.2019");
  test.equal(format.testables.time(1546297200000), "00:00");
  test.end();
});

tap.test("duration", test => {
  test.equal(format.duration(5 * 60 * 60 * 1000), "5h 00m");
  test.equal(format.duration(30 * 1000), "0h 00m");
  test.equal(format.duration(-30 * 1000), "0h 01m");
  test.equal(format.duration(90 * 1000), "0h 01m");
  test.equal(format.duration(-90 * 1000), "0h 02m");
  test.equal(format.duration(-600 * 1000), "0h 10m");
  test.end();
});

tap.test("entryList", test => {
  test.same(
    format.entryList({
      "1571004000": [
        { id: 105, from: 1571043673, to: 1571043674, valid: true },
        { id: 106, from: 1571043678, to: 1571044420, valid: true }
      ]
    }),
    [
      {
        date: "14.10.2019",
        valid: true,
        from: "11:01",
        to: "11:13",
        duration: "0h 12m",
        entries: [
          { from: "11:01", valid: true, to: "11:01", duration: "0h 00m" },
          { from: "11:01", valid: true, to: "11:13", duration: "0h 12m" }
        ]
      }
    ]
  );
  test.same(
    format.entryList({
      "1571004000": [
        { id: 105, from: 1571043673, to: 1571043674, valid: true },
        { id: 106, from: 1571043678, to: 1571044420, valid: false }
      ]
    }),
    [
      {
        date: "14.10.2019",
        valid: false,
        from: "11:01",
        to: "XX:XX",
        duration: "0h 00m",
        entries: [
          { from: "11:01", valid: true, to: "11:01", duration: "0h 00m" },
          { from: "11:01", valid: false, to: "XX:XX", duration: "Xh XXm" }
        ]
      }
    ]
  );
  test.same(
    format.entryList({
      "1571004000": [
        { id: 105, from: 1571043673, to: 1571043674, valid: false }
      ]
    }),
    [
      {
        date: "14.10.2019",
        valid: false,
        from: "11:01",
        to: "XX:XX",
        duration: "Xh XXm",
        entries: [
          { from: "11:01", valid: false, to: "XX:XX", duration: "Xh XXm" }
        ]
      }
    ]
  );
  test.end();
});
