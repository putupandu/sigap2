from fastapi import FastAPI
from pydantic import BaseModel
from transformers import pipeline
import uvicorn
import logging

# Konfigurasi Logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="SIGAP NLP Spam Filter",
    description="Microservice untuk mendeteksi spam/curhatan vs laporan darurat/bencana.",
    version="1.0.0",
)

# Model NLI Multilingual yang sangat baik untuk Zero-Shot Classification Bahasa Indonesia
MODEL_NAME = "MoritzLaurer/mDeBERTa-v3-base-mnli-xnli"

# Inisialisasi pipeline saat aplikasi berjalan
logger.info(f"Loading Zero-Shot Classification pipeline with model: {MODEL_NAME}")
# Menyimpan pipeline di memori
classifier = pipeline("zero-shot-classification", model=MODEL_NAME)
logger.info("Pipeline loaded successfully!")

# Label untuk Zero-Shot Classification
# Karena ini model NLI, deskripsi label yang panjang dan jelas akan lebih baik
CANDIDATE_LABELS = [
    "laporan bencana alam atau keadaan darurat", # Representasi VALID
    "pesan spam, curhatan pribadi, atau obrolan biasa" # Representasi SPAM
]

# Model Input API
class ClassifyRequest(BaseModel):
    text: str

# Model Output API
class ClassifyResponse(BaseModel):
    status: str # "VALID" atau "SPAM"
    confidence: float
    reasoning: dict # Optional: untuk melihat skor detail dari masing-masing label

@app.post("/v1/classify", response_model=ClassifyResponse)
async def classify_text(req: ClassifyRequest):
    text = req.text.strip()
    
    if not text:
        return ClassifyResponse(
            status="SPAM", 
            confidence=1.0,
            reasoning={"error": "Empty text"}
        )

    # Lakukan klasifikasi
    result = classifier(text, CANDIDATE_LABELS, multi_label=False)
    
    # Label dengan skor tertinggi (index 0)
    top_label = result['labels'][0]
    top_score = result['scores'][0]
    
    # Mapping dari label kalimat ke status singkat
    if top_label == CANDIDATE_LABELS[0]:
        status = "VALID"
    else:
        status = "SPAM"
        
    return ClassifyResponse(
        status=status,
        confidence=top_score,
        reasoning={
            result['labels'][0]: result['scores'][0],
            result['labels'][1]: result['scores'][1]
        }
    )

@app.get("/health")
async def health_check():
    return {"status": "ok", "model": MODEL_NAME}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
