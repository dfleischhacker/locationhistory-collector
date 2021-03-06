package rest

import "strings"

func GetIndexFile() []byte {
	return []byte(strings.Replace(`
<!DOCTYPE html>
<html>
<head>
    <meta charset=utf-8 />
    <title>GPX trackpoints and waypoints</title>
    <meta name='viewport' content='initial-scale=1,maximum-scale=1,user-scalable=no' />
    <script src='https://api.mapbox.com/mapbox.js/v3.2.0/mapbox.js'></script>
    <script src='https://code.jquery.com/jquery-3.5.1.min.js'></script>
    <link href='https://api.mapbox.com/mapbox.js/v3.2.0/mapbox.css' rel='stylesheet' />
    <style>
        body { margin:0; padding:0; }
        #map { position:absolute; top:0; bottom:0; width:100%; }
    </style>
</head>
<body>

<script src='https://api.mapbox.com/mapbox.js/plugins/leaflet-omnivore/v0.2.0/leaflet-omnivore.min.js'></script>

<div id='map'></div>

<script>
var token;
$(document).ready(function() {
    $.get("/token", function(data) {
        L.mapbox.accessToken = data;
        initMap();
    })
});

    function initMap() {
    var map = L.mapbox.map('map')
        .addLayer(L.mapbox.styleLayer('mapbox://styles/mapbox/streets-v11'));

    // omnivore will AJAX-request this file behind the scenes and parse it:
    // note that there are considerations:
    // - The file must either be on the same domain as the page that requests it,
    //   or both the server it is requested from and the user's browser must
    //   support CORS.
    var runLayer = omnivore.gpx('/locations/owntracks/daniel/70A43116-AF9A-4570-9040-9262AA75CCB9')
        .on('ready', function() {
            map.fitBounds(runLayer.getBounds());
			runLayer.eachLayer(function(layer) {
            // See the .bindPopup documentation for full details. This
            // dataset has a property called name: your dataset might not,
            // so inspect it and customize to taste.
            layer.bindPopup(layer.feature.properties.time);
        });
        })
        .addTo(map);
    }
</script>

</body>
</html>


`, "__TOKEN__", "", 1))
}
