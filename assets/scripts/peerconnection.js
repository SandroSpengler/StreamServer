const establishPeerConnection = async () => {
    let pc = new RTCPeerConnection();

    pc.ontrack = (event) => {
        if (event.track.kind == "video") {
            let element = document.createElement(event.track.kind);
            element.srcObject = event.streams[0];
            element.autoplay = true;
            element.controls = true;

            document.getElementById('rtmpFeed').appendChild(element)
        }
    }

    pc.addTransceiver("video");
    pc.addTransceiver("audio");

    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);

    try {
        // comment
        const req = await fetch("/createPeerConnection", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(offer),
        });

        const res = await req.json();

        await pc.setRemoteDescription(res)

        const placeHolder = document.getElementById("loadingPlaceholder")
        placeHolder.style.display = "none"
    } catch (e) {
        let loadingText = document.getElementById("loadingText")

        loadingText.innerHTML = "Error establishing PeerConnection";
        loadingText.style.color = "red";

        console.error(e);
    }
}


window.onload = async () => {
    await establishPeerConnection();
}
