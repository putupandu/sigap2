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

	scores, likely, _ := localClassifier.ProbScores(words)
	
	fmt.Printf("[ML] Analisa: %q\n", text)
	fmt.Printf("[ML] Irrelevant Prob: %.2f%%, Disaster Prob: %.2f%%\n", scores[0]*100, scores[1]*100)

	// likely == 0 berarti model memilih "Irrelevant"
	if likely == 0 && scores[0] > 0.60 {
		fmt.Println("[ML] REJECTED: Teks terdeteksi sebagai Irrelevant (>60%)")
		return true
	}

	// Atau jika probabilitas irrelevant di atas 70% walaupun bukan likely utama (edge case, meski jarang)
	if scores[0] > 0.70 {
		fmt.Println("[ML] REJECTED: Probabilitas Irrelevant sangat tinggi (>70%)")
		return true
	}

	fmt.Println("[ML] ACCEPTED: Lolos filter ML")
	return false
}
