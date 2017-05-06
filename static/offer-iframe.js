document.addEventListener("DOMContentLoaded", function() {

	window.addEventListener("message", function(event) {
		if (event.data.rpcId !== "0") {
			return;
		}
		if (event.data.error) {
			console.log("ERROR: " + event.data.error);
			return;
		}
		var elt = document.getElementById("offer-iframe");
		elt.setAttribute("src", event.data.uri);
	});

	window.parent.postMessage({
		renderTemplate: {
			rpcId: "0",
			template: '<script id="ssc-script" src="' + window.location.protocol +
					'//$API_HOST/.sandstorm-token/$API_TOKEN/static/embed.min.js' +
					'"></script>',
			forSharing: true,
			clipboardButton: 'left',
			// TODO: restrict roles
		}
	}, "*");

});
