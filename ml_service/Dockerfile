# Gunakan python 3.9 slim yang lebih ringan
FROM python:3.9-slim

# Set working directory di dalam container
WORKDIR /app

# Install system dependencies yang dibutuhkan (misal build-essential jika diperlukan oleh torch)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy file requirements terlebih dahulu untuk caching docker layer
COPY requirements.txt .

# Install PyTorch CPU version first to save 700MB of download and speed up the build
RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu --default-timeout=2000 --retries 10

# Install python packages
RUN pip install --no-cache-dir -r requirements.txt --default-timeout=2000 --retries 10

# Download pre-trained model saat proses build image agar saat container running lebih cepat
# Kita jalankan script python kecil untuk mengunduh model cache ke dalam image
RUN python -c "from transformers import pipeline; pipeline('zero-shot-classification', model='MoritzLaurer/mDeBERTa-v3-base-mnli-xnli')"

# Copy seluruh file script
COPY . .

# Expose port untuk FastAPI (port internal container)
EXPOSE 8000

# Jalankan server uvicorn
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
