document.addEventListener("DOMContentLoaded", function() {
    const btnSos = document.getElementById('btn-sos');
    const sosDesc = document.getElementById('sos-desc');
    const sosName = document.getElementById('sos-name');
    const sosStatus = document.getElementById('sos-status');
    const sosRipple = document.getElementById('sos-ripple');
    const needsOtherCheckbox = document.getElementById('needs-other-checkbox');
    const needsOtherInput = document.getElementById('sos-needs-other');

    if (!btnSos) return;

    // ============================================================
    // Checkbox toggle styling
    // ============================================================
    document.querySelectorAll('.needs-checkbox').forEach(cb => {
        cb.addEventListener('change', function() {
            const wrapper = this.closest('.needs-option').querySelector('div');
            if (this.checked) {
                wrapper.classList.remove('border-slate-200', 'bg-slate-50');
                wrapper.classList.add('border-red-500', 'bg-red-50', 'ring-2', 'ring-red-200');
                wrapper.querySelector('.text-sm').classList.remove('text-slate-700');
                wrapper.querySelector('.text-sm').classList.add('text-red-700', 'font-bold');
            } else {
                wrapper.classList.remove('border-red-500', 'bg-red-50', 'ring-2', 'ring-red-200');
                wrapper.classList.add('border-slate-200', 'bg-slate-50');
                wrapper.querySelector('.text-sm').classList.remove('text-red-700', 'font-bold');
                wrapper.querySelector('.text-sm').classList.add('text-slate-700');
            }
        });
    });

    // Show/hide "Lainnya" text input
    if (needsOtherCheckbox) {
        needsOtherCheckbox.addEventListener('change', function() {
            if (this.checked) {
                needsOtherInput.classList.remove('hidden');
                needsOtherInput.focus();
            } else {
                needsOtherInput.classList.add('hidden');
                needsOtherInput.value = '';
            }
        });
    }

    // ============================================================
    // Collect selected needs as comma-separated string
    // ============================================================
    function getSelectedNeeds() {
        const selected = [];
        document.querySelectorAll('.needs-checkbox:checked').forEach(cb => {
            if (cb.id === 'needs-other-checkbox') {
                const otherVal = needsOtherInput.value.trim();
                if (otherVal) {
                    selected.push(otherVal);
                }
            } else {
                selected.push(cb.value);
            }
        });
        return selected.join(', ');
    }

    // ============================================================
    // SOS Button Click
    // ============================================================
    btnSos.addEventListener('click', function() {
        // Show loading state
        btnSos.disabled = true;
        sosRipple.classList.remove('hidden');
        sosStatus.classList.remove('hidden', 'bg-red-50', 'text-red-700', 'bg-green-50', 'text-green-700');
        sosStatus.classList.add('bg-blue-50', 'text-blue-700');
        sosStatus.innerHTML = `
            <div class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-700" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Mendapatkan koordinat GPS Anda...
            </div>
        `;

        if (!navigator.geolocation) {
            showError("Geolocation tidak didukung oleh browser Anda.");
            return;
        }

        navigator.geolocation.getCurrentPosition(
            (position) => {
                sendSOS(position.coords.latitude, position.coords.longitude);
            },
            (error) => {
                showError("Gagal mendapatkan lokasi. Pastikan izin GPS diaktifkan.");
            },
            { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 }
        );
    });

    function sendSOS(lat, lng) {
        sosStatus.innerHTML = `
            <div class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-700" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Mengirim laporan ke server...
            </div>
        `;

        const formData = new FormData();
        formData.append('latitude', lat);
        formData.append('longitude', lng);
        formData.append('description', sosDesc.value);
        if (sosName) formData.append('reporter_name', sosName.value);
        formData.append('needs', getSelectedNeeds());

        fetch('/api/sos', {
            method: 'POST',
            body: formData,
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        })
        .then(response => {
            if (!response.ok) throw new Error("Server error");
            return response.json();
        })
        .then(data => {
            sosRipple.classList.add('hidden');
            sosStatus.classList.remove('bg-blue-50', 'text-blue-700');
            sosStatus.classList.add('bg-green-50', 'text-green-700');
            sosStatus.innerHTML = `
                <div class="flex flex-col items-center">
                    <svg class="w-8 h-8 text-green-500 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                    <span class="font-medium">Sinyal SOS Berhasil Dikirim!</span>
                    <span class="text-sm mt-1">Tim relawan terdekat telah menerima notifikasi. Tetap tenang dan tunggu bantuan.</span>
                </div>
            `;
            // Redirect after 3 seconds back to login page
            setTimeout(() => {
                window.location.href = "/login";
            }, 3000);
        })
        .catch(error => {
            showError("Terjadi kesalahan saat mengirim data. Coba lagi.");
        });
    }

    function showError(msg) {
        btnSos.disabled = false;
        sosRipple.classList.add('hidden');
        sosStatus.classList.remove('bg-blue-50', 'text-blue-700');
        sosStatus.classList.add('bg-red-50', 'text-red-700');
        sosStatus.innerHTML = `
            <div class="flex items-center justify-center">
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                ${msg}
            </div>
        `;
    }
});
