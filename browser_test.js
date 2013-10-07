casper.test.begin('Existing resource', 9, function suite(test) {
  casper.start('http://localhost:8080/resource/tnr_1140686', function() {
    test.assertHttpStatus(200, 'response status code 200');
    test.assertTitle('Azur', 'title as expected');
  });

  casper.thenOpen('http://localhost:8080/resource/tnr_1140686.html', function(response) {
    test.assertHttpStatus(200, 'response status code 200');
    test.assertMatch(response.headers.get('Content-Type'), /^text\/html/, 'correct content-type');
    test.assertTitle('Azur', 'title as expected');
  });

  casper.thenOpen('http://localhost:8080/resource/tnr_1140686.json', function(response) {
    test.assertMatch(response.headers.get('Content-Type'), /^application\/json/, 'correct content-type');
  });

  casper.thenOpen('http://localhost:8080/resource/tnr_1140686.rdf', function(response) {
    test.assertMatch(response.headers.get('Content-Type'), /^application\/x-trig/, 'correct content-type');
  });

  casper.thenOpen('http://localhost:8080/resource/tnr_1140686.zappa', function() {
    test.assertHttpStatus(400, 'status 400 on bad request');
    this.test.assertTextExists('Unsupported output format', 'error message in page body');
  });

  casper.run(function() {
    test.done();
  });
});

casper.test.begin('Missing resource', 1, function suite(test) {
  casper.start('http://localhost:8080/an/very/unlikely/path/doh', function() {
    test.assertHttpStatus(404, 'response status code 404');
  }).run(function() {
    test.done();
  });
});