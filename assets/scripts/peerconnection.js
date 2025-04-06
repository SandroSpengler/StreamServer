const establishPeerConnection = async () => {
	const config = {
		iceServers: [
			{ urls: "stun:stun.l.google.com:19302" },
			// correct credentials from the docker-compose.yaml
			{
				urls: "turn:axfert.com",
				username: "your_username",
				credential: "your_credential",
			},
		],
	};

	let peerConnection = new RTCPeerConnection(config);

	peerConnection.ontrack = (event) => {
		if (event.track.kind == "video") {
			let element = document.createElement(event.track.kind);
			element.srcObject = event.streams[0];
			element.autoplay = true;
			element.controls = true;

			document.getElementById("rtmpFeed").appendChild(element);
		}
	};

	peerConnection.onconnectionstatechange = () => {
		console.log("Connection state:", peerConnection.connectionState);

		if (peerConnection.connectionState == "connecting") {
			Toastify({
				text: "Connecting to the server...",
				duration: 3000,
				gravity: "top", // `top` or `bottom`
				close: true,
				position: "right", // `left`, `center` or `right`
				stopOnFocus: true, // Prevents dismissing of toast on hover
				onClick: function () {}, // Callback after click
			}).showToast();
		}

		if (peerConnection.connectionState == "connected") {
			Toastify({
				text: "Connected!",
				duration: 3000,
				gravity: "top", // `top` or `bottom`
				close: true,
				position: "right", // `left`, `center` or `right`
				stopOnFocus: true, // Prevents dismissing of toast on hover
				style: {
					background: "linear-gradient(to right, #00b09b, #96c93d)",
				},
				onClick: function () {}, // Callback after click
			}).showToast();
		}

		if (peerConnection.connectionState == "closed") {
			Toastify({
				text: "Connection closed!",
				duration: 3000,
				gravity: "top", // `top` or `bottom`
				close: true,
				position: "right", // `left`, `center` or `right`
				stopOnFocus: true, // Prevents dismissing of toast on hover
				style: {
					background: "red",
				},
				onClick: function () {}, // Callback after click
			}).showToast();
		}
	};

	peerConnection.addTransceiver("video");
	peerConnection.addTransceiver("audio");

	const offer = await peerConnection.createOffer();
	await peerConnection.setLocalDescription(offer);

	try {
		const req = await fetch("/createPeerConnection", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(offer),
		});

		const res = await req.json();
		await peerConnection.setRemoteDescription(res);

		const placeHolder = document.getElementById("loadingPlaceholder");
		placeHolder.style.display = "none";
	} catch (e) {
		let loadingText = document.getElementById("loadingText");
		loadingText.innerHTML = "Error establishing PeerConnection";
		loadingText.style.color = "red";

		console.error(e);
	}
};

window.onload = async () => {
	await establishPeerConnection();

	if (window.pc && window.pc.connectionState !== "closed") {
		window.pc.close();
	}
};
