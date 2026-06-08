document.addEventListener("DOMContentLoaded", function () {
    const mapEl = document.getElementById('map');
    if (!mapEl) return;

    // Base Maps
    const satellite = L.tileLayer('https://{s}.google.com/vt/lyrs=s,h&x={x}&y={y}&z={z}', {
        maxZoom: 20,
        subdomains: ['mt0', 'mt1', 'mt2', 'mt3'],
        attribution: '&copy; Google Maps'
    });

    const osm = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors'
    });

    const streetMaps = L.tileLayer('https://{s}.google.com/vt/lyrs=m&x={x}&y={y}&z={z}', {
        maxZoom: 20,
        subdomains: ['mt0', 'mt1', 'mt2', 'mt3'],
        attribution: '&copy; Google Maps'
    });

    // Initialize map centered on Indonesia
    const map = L.map('map', {
        center: [-0.789275, 113.921327],
        zoom: 5,
        layers: [satellite] // Default layer
    });

    // Overlay Layer Groups
    const dataBencanaLayer = L.layerGroup();
    const gempaTerbaruLayer = L.layerGroup(); // For autogempa
    const gempaRealtimeLayer = L.layerGroup();
    const gempaDirasakanLayer = L.layerGroup();

    // Add default overlays to map
    dataBencanaLayer.addTo(map);
    gempaTerbaruLayer.addTo(map);
    gempaRealtimeLayer.addTo(map);
    gempaDirasakanLayer.addTo(map);

    const baseMaps = {
        "Satellite": satellite,
        "Osm": osm,
        "Street Maps": streetMaps
    };

    const overlayMaps = {
        "Data Bencana": dataBencanaLayer,
        "Gempa Terbaru (Auto)": gempaTerbaruLayer,
        "Gempa Terkini": gempaRealtimeLayer,
        "Gempa Dirasakan": gempaDirasakanLayer
    };

    // Add Layer Control
    L.control.layers(baseMaps, overlayMaps).addTo(map);

    // 1. Fetch markers from API (Data Bencana SIGAP)
    fetch('/api/reports/markers')
        .then(response => response.json())
        .then(data => {
            if (!data) return;

            data.forEach(marker => {
                let color = marker.status === 'pending' ? 'red' : marker.status === 'process' ? 'orange' : 'green';

                const markerHtml = `<div style="background-color: ${color}; width: 24px; height: 24px; border-radius: 50%; border: 3px solid white; box-shadow: 0 2px 5px rgba(0,0,0,0.5);"></div>`;
                const customIcon = L.divIcon({
                    html: markerHtml,
                    className: '',
                    iconSize: [24, 24],
                    iconAnchor: [12, 12]
                });

                L.marker([marker.lat, marker.lng], { icon: customIcon })
                    .bindPopup(`
                        <div class="text-sm">
                            <p class="font-bold mb-1">${marker.name}</p>
                            ${marker.needs ? `<p class="mb-1 text-xs text-slate-500 line-clamp-1">Kebutuhan: ${marker.needs}</p>` : ''}
                            <p class="mb-2">Status: <span class="uppercase text-xs font-semibold">${marker.status}</span></p>
                            <a href="/reports/${marker.id}" class="text-sigap-600 hover:underline">Lihat Detail</a>
                        </div>
                    `)
                    .addTo(dataBencanaLayer);
            });
        })
        .catch(err => console.error("Error fetching map markers:", err));

    // Helper to fetch BMKG Data
    function fetchBMKG(url, layerGroup, isAutoGempa = false) {
        fetch(url)
            .then(res => res.json())
            .then(data => {
                let gempas = [];
                if (data && data.Infogempa && data.Infogempa.gempa) {
                    if (Array.isArray(data.Infogempa.gempa)) {
                        gempas = data.Infogempa.gempa;
                    } else {
                        gempas = [data.Infogempa.gempa];
                    }
                }

                if (gempas.length > 0) {
                    layerGroup.clearLayers();
                }

                gempas.forEach(g => {
                    if (!g.Coordinates) return;
                    let coords = g.Coordinates.split(',');
                    let lat = parseFloat(coords[0]);
                    let lng = parseFloat(coords[1]);
                    let mag = parseFloat(g.Magnitude);

                    // Style marker based on magnitude
                    let color = mag >= 5 ? '#ef4444' : (mag >= 4 ? '#eab308' : '#22c55e'); // Red, Yellow, Green
                    
                    let extraStyle = '';
                    if (isAutoGempa) {
                        extraStyle = `animation: pulse 2s infinite; box-shadow: 0 0 20px ${color}; border-width: 3px; z-index: 1000;`;
                        color = '#dc2626'; // Force red for auto gempa
                    } else {
                        extraStyle = `box-shadow: 0 0 10px ${color};`;
                    }

                    const markerHtml = `
                        <div style="background-color: ${color}; opacity: 0.9; width: 30px; height: 30px; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: white; font-size: 11px; font-weight: bold; border: 2px solid white; ${extraStyle}">
                            ${mag}
                        </div>
                    `;
                    const icon = L.divIcon({
                        html: markerHtml,
                        className: '',
                        iconSize: [30, 30],
                        iconAnchor: [15, 15]
                    });

                    L.marker([lat, lng], { icon: icon })
                        .bindPopup(`
                            <div class="text-sm min-w-[200px]">
                                <p class="font-bold border-b pb-1 mb-1 text-slate-800">${isAutoGempa ? '🚨 GEMPA TERBARU M' : 'Gempa Bumi M'} ${g.Magnitude}</p>
                                <p class="mb-1 text-slate-600"><b>Waktu:</b> ${g.Tanggal} ${g.Jam}</p>
                                <p class="mb-1 text-slate-600"><b>Kedalaman:</b> ${g.Kedalaman}</p>
                                <p class="mb-1 text-slate-600"><b>Wilayah:</b> ${g.Wilayah}</p>
                                ${g.Potensi ? `<p class="mb-1 text-red-600 font-semibold">${g.Potensi}</p>` : ''}
                                ${g.Dirasakan ? `<p class="mb-1 text-orange-600 font-semibold"><b>Dirasakan:</b> ${g.Dirasakan}</p>` : ''}
                                <p class="text-xs text-slate-400 mt-2 text-right">Sumber: BMKG</p>
                            </div>
                        `)
                        .addTo(layerGroup);
                });
            })
            .catch(err => console.error("Error fetching BMKG data:", err));
    }

    // Function to load all BMKG data
    function updateBMKGData() {
        // 1. Fetch BMKG Auto Gempa (The absolute latest single earthquake)
        fetchBMKG('https://data.bmkg.go.id/DataMKG/TEWS/autogempa.json', gempaTerbaruLayer, true);
        
        // 2. Fetch BMKG Gempa Terkini
        fetchBMKG('https://data.bmkg.go.id/DataMKG/TEWS/gempaterkini.json', gempaRealtimeLayer);

        // 3. Fetch BMKG Gempa Dirasakan
        fetchBMKG('https://data.bmkg.go.id/DataMKG/TEWS/gempadirasakan.json', gempaDirasakanLayer);
    }

    // Load immediately
    updateBMKGData();
    
    // Auto-update every 60 seconds (Real-time polling)
    setInterval(updateBMKGData, 60000);
});
