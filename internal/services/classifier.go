package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrukh/bayesian"
)

const (
	ClassIrrelevant bayesian.Class = "Irrelevant"
	ClassDisaster   bayesian.Class = "Disaster"
)

var localClassifier *bayesian.Classifier

type Dataset struct {
	Irrelevant []string `json:"irrelevant"`
	Disaster   []string `json:"disaster"`
}

func InitMLClassifier() error {
	localClassifier = bayesian.NewClassifier(ClassIrrelevant, ClassDisaster)

	// Path to dataset (assuming working directory is root of project)
	datasetPath := filepath.Join("internal", "services", "data", "dataset_nlp.json")
	file, err := os.ReadFile(datasetPath)
	if err != nil {
		fmt.Printf("[ML] WARNING: could not read dataset file: %v\n", err)
		return err
	}

	var data Dataset
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Printf("[ML] WARNING: could not parse dataset JSON: %v\n", err)
		return err
	}

	for _, text := range data.Irrelevant {
		words := tokenizeText(text)
		if len(words) > 0 {
			localClassifier.Learn(words, ClassIrrelevant)
		}
	}

	for _, text := range data.Disaster {
		words := tokenizeText(text)
		if len(words) > 0 {
			localClassifier.Learn(words, ClassDisaster)
		}
	}

	fmt.Printf("[ML] Classifier initialized successfully with %d irrelevant and %d disaster patterns.\n", len(data.Irrelevant), len(data.Disaster))
	return nil
}

func tokenizeText(text string) []string {
	text = strings.ToLower(text)
	replacer := strings.NewReplacer(".", " ", ",", " ", "!", " ", "?", " ", "-", " ", "\n", " ", "\r", " ")
	text = replacer.Replace(text)
	return strings.Fields(text)
}

// emergencyKeywords are CONTEXTUAL phrases that strongly indicate a real disaster/emergency.
// IMPORTANT: We use multi-word phrases instead of single words to avoid false positives.
// e.g. "terjebak" alone could match "terjebak macet" (traffic) which is NOT a disaster.
var emergencyKeywords = []string{
	// Terjebak/terperangkap — only with disaster context
	"terjebak banjir", "terjebak longsor", "terjebak gempa", "terjebak kebakaran",
	"terjebak reruntuhan", "terjebak tanah longsor",
	"terperangkap banjir", "terperangkap longsor", "terperangkap kebakaran",
	"terperangkap reruntuhan",
	"tertimbun longsor", "tertimbun reruntuhan", "tertimbun tanah",
	"tertimpa reruntuhan", "tertimpa bangunan", "tertimpa pohon",
	// Bencana alam — these are inherently disaster-related
	"tanah longsor", "banjir bandang", "banjir merendam",
	"gempa bumi", "tsunami",
	"kebakaran", "terbakar",
	"angin puting beliung", "angin kencang",
	"pohon tumbang",
	// Situasi darurat spesifik bencana
	"tenggelam", "hanyut terbawa",
	"reruntuhan bangunan", "bangunan runtuh", "rumah roboh",
	// Korban/evakuasi
	"evakuasi korban", "butuh evakuasi", "perlu evakuasi",
	"pengungsian", "pengungsi", "posko bencana", "posko pengungsian",
	"korban bencana", "korban banjir", "korban longsor", "korban gempa",
	// Kebutuhan darurat bencana
	"butuh bantuan darurat", "minta bantuan evakuasi",
	"butuh makanan", "butuh air bersih", "butuh obat", "butuh selimut", "butuh tenda",
	"terisolasi banjir", "terisolasi longsor",
}

// containsEmergencyKeyword checks if text contains any emergency-related keyword
func containsEmergencyKeyword(text string) bool {
	lower := strings.ToLower(text)
	for _, kw := range emergencyKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// CheckIfIrrelevantML returns true if the machine learning model classifies the text as a non-disaster (irrelevant)
func CheckIfIrrelevantML(text string) bool {
	if localClassifier == nil {
		fmt.Println("[ML] Classifier not initialized, skipping ML check")
		return false
	}

	words := tokenizeText(text)
	if len(words) == 0 {
		return true // empty text is irrelevant
	}

	// SAFETY NET: If text contains any emergency/disaster keyword, ALWAYS let it through.
	// This prevents false positives on legitimate disaster reports.
	if containsEmergencyKeyword(text) {
		fmt.Printf("[ML] BYPASS: Teks mengandung kata darurat/bencana, langsung lolos. Text: %q\n", text)
		return false
	}

	scores, likely, _ := localClassifier.ProbScores(words)

	fmt.Printf("[ML] Analisa: %q\n", text)
	fmt.Printf("[ML] Irrelevant Prob: %.2f%%, Disaster Prob: %.2f%%\n", scores[0]*100, scores[1]*100)

	// likely == 0 berarti model memilih "Irrelevant" dan probabilitas harus tinggi
	if likely == 0 && scores[0] >= 0.75 {
		fmt.Println("[ML] REJECTED: Teks terdeteksi sebagai Irrelevant (>=75%)")
		return true
	}

	fmt.Println("[ML] ACCEPTED: Lolos filter ML")
	return false
}
