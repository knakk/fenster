casper.test.begin('Existing resource', 4, function suite(test) {
  casper.start("http://localhost:8080/resource/tnr_1140686", function() {
    test.assertHttpStatus(200, "response status code 200");
    test.assertTitle("Azur", "title as expected");
  });

  casper.thenOpen("http://localhost:8080/resource/tnr_1140686.html", function() {
    test.assertHttpStatus(200, "response status code 200");
    test.assertTitle("Azur", "title as expected");
  });

  casper.run(function () {
    test.done();
  });
});

casper.test.begin('Missing resource', 1, function suite(test) {
  casper.start("http://localhost:8080/an/very/unlikely/path/doh", function() {
    test.assertHttpStatus(404, "response status code 404");
  }).run(function () {
    test.done();
  });
});