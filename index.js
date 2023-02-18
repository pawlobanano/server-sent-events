var source = new EventSource("http://localhost:3000/stream?topic=topic1");
console.log(source)
source.onmessage = function(event) {
  var data = JSON.parse(event.data);
  console.log("Received event: "+ event + " from topic: " + data.topic + " with data: " + data.data);
};
