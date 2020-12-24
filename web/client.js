
const {DummyRequest, SimpleResponse} = require('./gen/protos/config_pb.js');
const {MyRPCServerClient} = require('./gen/protos/config_grpc_web_pb.js');

var client = new MyRPCServerClient(location.protocol + "//" + location.host);

var request = new DummyRequest();

client.dummy(request, {}, (err, response) => {
	console.log(err);
	console.log(response);

	if (response != null) {
		document.getElementById("myheader").innerHTML = response.getMessage();
	}
});