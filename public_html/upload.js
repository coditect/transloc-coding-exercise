window.addEventListener("load", function() {

	var formElement = document.getElementById("upload"),
		inputElement = document.getElementById("file"),
		infoElement = document.querySelector(".upload-info"),
		originalInfo = infoElement.innerHTML;

	function uploadFile(file) {
		var req = new XMLHttpRequest();
		req.open("POST", "/geoip");
		req.onload = function() {
			inputElement.disabled = false;

			if (req.status == 200) {
				updateHeatmap();
				setInfo(originalInfo, false);
			} else {
				setInfo(req.responseText, true);
			}
		};
		req.send(new FormData(formElement));
		setInfo("Loading…", false);
		inputElement.disabled = true;
	}

	function setInfo(content, isError) {
		infoElement.innerHTML = content;
		infoElement.classList.toggle("error", isError);
	}

	inputElement.addEventListener("change", function(event) {
		if (inputElement.files.length > 0) {
			uploadFile(inputElement.files[0]);
		}
	}, false);

}, false);
