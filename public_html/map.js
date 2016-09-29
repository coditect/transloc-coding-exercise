window.addEventListener("load", function() {

	var map = L.map('map'),
		heatmap = null,
		timeout = null,
		gradient = {0.0: '#00f', 0.05: '#0ff', 0.1: '#0f0', 0.2: '#ff0', 0.4: '#f00', 0.8: '#f0f'};

	function updateHeatmap(event) {
		if (timeout != null) {
			clearTimeout(timeout);
		}
		timeout = setTimeout(updateHeatmapNow, 250);
	}

	function updateHeatmapNow() {
		var bounds = map.getBounds();
		var req = new XMLHttpRequest();
		req.open("GET", "/geoip?north=" + bounds.getNorth() + "&south=" + bounds.getSouth() + "&east=" + bounds.getEast() + "&west=" + bounds.getWest())
		req.onload = function() {
			var data = JSON.parse(req.responseText);

			if (heatmap != null) {
				map.removeLayer(heatmap);
			}

			var radius = 2 * Math.pow(map.getZoom() + 1, 1.25);
			heatmap = L.heatLayer(data, {
				radius: radius,
				blur: radius * 0.75,
				max: 33,
				gradient: gradient,
			}).addTo(map);

		};
		req.onerror = function() {
			console.error("Unable to load heatmap data: " + req.responseText);
		};
		req.send();
	}

	L.tileLayer('http://server.arcgisonline.com/ArcGIS/rest/services/Canvas/World_Light_Gray_Base/MapServer/tile/{z}/{y}/{x}', {
		attribution: 'Tiles &copy; Esri &mdash; Esri, DeLorme, NAVTEQ',
		maxZoom: 16
	}).addTo(map);

	map.on('zoomend', updateHeatmap);
	map.on('dragend', updateHeatmap);
	map.setView([15, 0], 2);

	window.updateHeatmap = updateHeatmapNow;

}, false);
