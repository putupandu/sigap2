package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type NLPResult struct {
	IsDisasterRelated bool      `json:"is_disaster_related"`
	Reason            string    `json:"reason"`
	Urgency           string    `json:"urgency"`
	Items             []NLPItem `json:"items"`
}

type NLPItem struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

var (
	urlRegex       = regexp.MustCompile(`(?i)(https?://|www\.)[^\s]+`)
	gibberishRegex = regexp.MustCompile(`(?i)(wkwk|hahaha|asdf|qwer|zxcv|xixi|hehe)\w*`)
	spamWordRegex  = regexp.MustCompile(`(?i)\b(tes|test|coba|testing|halo|ping|p|anjing|babi|bangsat|kontol|memek|jembut|ngentot|tai|goblok|tolol|bajingan|asu|mabar|nongkrong|ngopi|healing|skripsi|tugas kuliah|dosen)\b`)
)

func detectSpamLeksikal(description, needs string) bool {
	combined := strings.ToLower(description + " " + needs)

	if urlRegex.MatchString(combined) {
		return true
	}
	if gibberishRegex.MatchString(combined) {
		return true
	}
	if spamWordRegex.MatchString(combined) {
		return true
	}

	return false
}

type MLClassifyRequest struct {
	Text string `json:"text"`
}

type MLClassifyResponse struct {
	Status     string  `json:"status"`
	Confidence float64 `json:"confidence"`
}

func checkSpamML(description, needs string) bool {
	combined := strings.TrimSpace(description + " " + needs)
	if combined == "" {
		return true
	}

	reqBody, _ := json.Marshal(MLClassifyRequest{Text: combined})

	mlURL := os.Getenv("ML_CLASSIFIER_URL")
	if mlURL == "" {
		// Default internal Docker network URL
		mlURL = "http://ml-classifier:8000/v1/classify"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(mlURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("[NLP] ML Classifier request error: %v, falling back to lexical\n", err)
		return detectSpamLeksikal(description, needs)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[NLP] ML Classifier returned status %d, falling back to lexical\n", resp.StatusCode)
		return detectSpamLeksikal(description, needs)
	}

	var mlResp MLClassifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&mlResp); err != nil {
		fmt.Printf("[NLP] ML Classifier decode error: %v, falling back to lexical\n", err)
		return detectSpamLeksikal(description, needs)
	}

	fmt.Printf("[NLP] ML Classifier result: %s (confidence: %.2f) for text: %q\n", mlResp.Status, mlResp.Confidence, combined)
	
	return mlResp.Status == "SPAM"
}

// detectUrgencyByKeywords uses keyword matching to determine urgency level.
// This serves as a reliable fallback when AI analysis fails or returns incorrect results.
// NOTE: Irrelevant detection is now handled by the ML Classifier (classifier.go).
func detectUrgencyByKeywords(description string, needs string) string {
	combined := strings.ToLower(description + " " + needs)


	// CRITICAL keywords: life-threatening, trapped, need immediate rescue
	criticalKeywords := []string{
		"terjebak", "terperangkap", "tertimbun", "tertimpa",
		"longsor", "tanah longsor",
		"tenggelam", "hanyut",
		"kebakaran", "terbakar", "api",
		"gempa", "reruntuhan", "runtuh",
		"tsunami",
		"luka parah", "luka berat", "patah tulang", "pendarahan", "berdarah",
		"tidak sadarkan diri", "pingsan", "kritis",
		"meninggal", "tewas",
		"tolong", "selamatkan", "evakuasi",
		"nyawa", "sekarat", "darurat",
		"banjir bandang",
		"anak kecil terjebak", "bayi",
	}

	for _, kw := range criticalKeywords {
		if strings.Contains(combined, kw) {
			return "critical"
		}
	}

	// HIGH keywords: serious but not immediately life-threatening
	highKeywords := []string{
		"banjir", "terendam", "genangan tinggi",
		"rumah rusak", "rumah roboh", "atap roboh",
		"akses terputus", "jalan terputus", "terisolasi",
		"kelaparan", "kedinginan", "dehidrasi",
		"lansia", "manula", "orang tua",
		"hamil", "ibu hamil",
		"luka ringan", "demam tinggi", "sakit parah",
		"tidak ada makanan", "tidak ada air",
		"mengungsi", "pengungsi", "pengungsian",
		"pohon tumbang",
	}

	for _, kw := range highKeywords {
		if strings.Contains(combined, kw) {
			return "high"
		}
	}

	return "" // no keyword match, let AI decide
}

// extractItemsByRegex is a fallback extractor that uses pattern matching to find
// items and quantities from informal Indonesian text.
// It handles patterns like:
// - "air mineral 10" / "10 air mineral"
// - "makanan 5 bungkus" / "5 bungkus makanan"
// - "butuh baju 5, selimut 3"
// - "saya membutuhkan air mineral 10, makanan 10, dan pakaian 5"
func extractItemsByRegex(text string) []NLPItem {
	lower := strings.ToLower(strings.TrimSpace(text))

	// Known item keywords that we can recognize in text
	knownItems := []struct {
		keywords []string // possible text mentions
		name     string   // normalized item name
	}{
		{[]string{"air mineral", "air minum", "air bersih", "air galon", "galon air"}, "air mineral"},
		{[]string{"nasi bungkus", "nasi kotak", "nasi box"}, "nasi bungkus"},
		{[]string{"mie instan", "mi instan", "mie", "indomie"}, "mie instan"},
		{[]string{"makanan", "makan"}, "makanan"},
		{[]string{"minuman", "minum"}, "minuman"},
		{[]string{"pakaian", "baju", "kaos"}, "pakaian"},
		{[]string{"celana"}, "celana"},
		{[]string{"selimut"}, "selimut"},
		{[]string{"obat-obatan", "obat p3k", "obat", "p3k"}, "obat-obatan"},
		{[]string{"masker"}, "masker"},
		{[]string{"sabun"}, "sabun"},
		{[]string{"pampers", "popok", "diapers"}, "popok/pampers"},
		{[]string{"susu formula", "susu bayi", "susu"}, "susu"},
		{[]string{"tenda"}, "tenda"},
		{[]string{"terpal", "terpal plastik"}, "terpal"},
		{[]string{"kasur", "matras", "alas tidur"}, "kasur/matras"},
		{[]string{"bantal"}, "bantal"},
		{[]string{"tikar"}, "tikar"},
		{[]string{"perahu karet", "perahu"}, "perahu karet"},
		{[]string{"genset", "generator"}, "genset"},
		{[]string{"senter", "lampu senter"}, "senter"},
		{[]string{"lilin"}, "lilin"},
		{[]string{"korek api", "korek"}, "korek api"},
		{[]string{"kompor"}, "kompor"},
		{[]string{"gas", "tabung gas", "elpiji"}, "tabung gas"},
		{[]string{"ember"}, "ember"},
		{[]string{"jerigen"}, "jerigen"},
		{[]string{"tali"}, "tali"},
		{[]string{"sarung"}, "sarung"},
		{[]string{"handuk"}, "handuk"},
		{[]string{"sandal", "sepatu"}, "alas kaki"},
		{[]string{"jas hujan", "raincoat"}, "jas hujan"},
		{[]string{"jaket"}, "jaket"},
		{[]string{"vitamin"}, "vitamin"},
		{[]string{"roti"}, "roti"},
		{[]string{"biskuit"}, "biskuit"},
		{[]string{"kopi"}, "kopi"},
		{[]string{"gula"}, "gula"},
		{[]string{"beras"}, "beras"},
	}

	var items []NLPItem
	foundItems := make(map[string]bool)

	// Pattern 1: "ITEM ANGKA" (e.g., "air mineral 10", "makanan 5")
	// Pattern 2: "ANGKA ITEM" (e.g., "10 air mineral", "5 makanan")
	// Pattern 3: "ITEM ANGKA UNIT" (e.g., "baju 5 pasang", "air 10 liter")
	// Pattern 4: "ANGKA UNIT ITEM" (e.g., "5 pasang baju", "10 liter air")

	units := []string{"pasang", "buah", "lembar", "helai", "bungkus", "bks",
		"botol", "btl", "liter", "lt", "kardus", "dus", "karung",
		"kg", "kilogram", "pack", "pcs", "unit", "set", "kotak",
		"boks", "box", "sachet", "sak", "batang", "roll", "gulung"}
	unitPattern := strings.Join(units, "|")

	for _, item := range knownItems {
		for _, keyword := range item.keywords {
			if !strings.Contains(lower, keyword) {
				continue
			}
			if foundItems[item.name] {
				continue
			}

			// Try Pattern 1: "keyword ANGKA [UNIT]"
			p1 := regexp.MustCompile(regexp.QuoteMeta(keyword) + `\s+(\d+)\s*(` + unitPattern + `)?`)
			if m := p1.FindStringSubmatch(lower); m != nil {
				qty := m[1]
				if m[2] != "" {
					qty += " " + m[2]
				}
				items = append(items, NLPItem{Name: item.name, Quantity: qty})
				foundItems[item.name] = true
				continue
			}

			// Try Pattern 2: "ANGKA [UNIT] keyword"
			p2 := regexp.MustCompile(`(\d+)\s*(` + unitPattern + `)?\s*` + regexp.QuoteMeta(keyword))
			if m := p2.FindStringSubmatch(lower); m != nil {
				qty := m[1]
				if m[2] != "" {
					qty += " " + m[2]
				}
				items = append(items, NLPItem{Name: item.name, Quantity: qty})
				foundItems[item.name] = true
				continue
			}

			// No quantity found but keyword exists — still add with "?" quantity
			// This handles abstract mentions like "butuh makanan" without specific numbers
			items = append(items, NLPItem{Name: item.name, Quantity: "sesuai kebutuhan"})
			foundItems[item.name] = true
		}
	}

	// Fallback: try to find ANY "number + word" or "word + number" patterns
	// for items not in the known list
	if len(items) == 0 {
		// Pattern: find all "ANGKA KATA" or "KATA ANGKA" pairs
		genericPattern := regexp.MustCompile(`(\d+)\s+([a-zA-Z\s]{2,20}?)(?:\s*[,.]|\s+dan\s|\s+\d|$)`)
		matches := genericPattern.FindAllStringSubmatch(lower, -1)
		for _, m := range matches {
			name := strings.TrimSpace(m[2])
			qty := m[1]
			// Filter out non-item words
			skipWords := []string{"orang", "tahun", "hari", "jam", "menit", "meter", "km", "rumah", "jiwa", "kepala", "kk"}
			skip := false
			for _, sw := range skipWords {
				if strings.Contains(name, sw) {
					skip = true
					break
				}
			}
			if !skip && len(name) > 1 && !foundItems[name] {
				// Clean trailing common words
				name = strings.TrimRightFunc(name, unicode.IsSpace)
				name = strings.TrimSuffix(name, " dan")
				name = strings.TrimSuffix(name, " yang")
				name = strings.TrimSuffix(name, " saya")
				name = strings.TrimSpace(name)
				if len(name) > 1 {
					items = append(items, NLPItem{Name: name, Quantity: qty})
					foundItems[name] = true
				}
			}
		}
	}

	return items
}

// AnalyzeReport uses Gemini API to extract urgency and structured items from text.
// Falls back to keyword-based detection if AI fails or returns a lower urgency than keywords suggest.
// Also falls back to regex-based item extraction if AI returns empty items.
func AnalyzeReport(description string, needs string) (string, string, error) {
	// Step 0: Check spam using ML Classifier (with lexical fallback)
	if checkSpamML(description, needs) {
		fmt.Printf("[NLP] SPAM detected for desc=%q needs=%q\n", description, needs)
		return "irrelevant", "[]", nil
	}

	// Step 1: Always run keyword-based detection first as baseline
	keywordUrgency := detectUrgencyByKeywords(description, needs)
	fmt.Printf("[NLP] Keyword detection result: %q for desc=%q needs=%q\n", keywordUrgency, description, needs)

	// Step 2: Always run regex-based item extraction as fallback baseline
	combinedForRegex := description
	if needs != "" {
		combinedForRegex += " " + needs
	}
	regexItems := extractItemsByRegex(combinedForRegex)
	fmt.Printf("[NLP] Regex extraction found %d items\n", len(regexItems))
	for _, item := range regexItems {
		fmt.Printf("[NLP]   - %s: %s\n", item.Name, item.Quantity)
	}

	// Step 3: Try AI analysis
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("[NLP] No GEMINI_API_KEY set, using keyword + regex fallback")
		finalUrgency := keywordUrgency
		if finalUrgency == "" {
			finalUrgency = "medium"
		}
		itemsJSON, _ := json.Marshal(regexItems)
		return finalUrgency, string(itemsJSON), nil
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Printf("[NLP] Gemini client error: %v, using regex fallback\n", err)
		finalUrgency := keywordUrgency
		if finalUrgency == "" {
			finalUrgency = "medium"
		}
		itemsJSON, _ := json.Marshal(regexItems)
		return finalUrgency, string(itemsJSON), err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	model.ResponseMIMEType = "application/json"

	// Gabungkan description dan needs menjadi satu teks untuk analisis yang lebih baik
	combinedText := strings.TrimSpace(description)
	if strings.TrimSpace(needs) != "" {
		combinedText = combinedText + "\nKebutuhan tambahan: " + strings.TrimSpace(needs)
	}

	prompt := fmt.Sprintf(`Kamu adalah AI SATPAM SUPER KETAT untuk sistem tanggap darurat bencana Indonesia (SIGAP).
Aplikasi ini HANYA untuk melaporkan BENCANA ALAM dan KEADAAN DARURAT.

=== TUGAS UTAMA: VALIDASI KONTEKS BENCANA ===
Langkah 1: Tentukan apakah teks ini BENAR-BENAR berhubungan dengan bencana/keadaan darurat.
- Set "is_disaster_related" = true HANYA jika teks menyebutkan konteks bencana alam (banjir, gempa, longsor, kebakaran, tsunami, angin puting beliung), kecelakaan berat, situasi darurat di posko/pengungsian, atau permintaan bantuan logistik untuk korban bencana.
- Set "is_disaster_related" = false jika teks adalah: curhatan pribadi, keluhan sehari-hari (lapar biasa, panas, capek, bosan), pesan iseng/spam, tes sistem, kata-kata kasar, obrolan biasa, atau APAPUN yang tidak ada hubungannya dengan bencana/darurat.
- WAJIB isi "reason" dengan penjelasan singkat mengapa kamu memutuskan true/false.

Langkah 2: Jika is_disaster_related = false, WAJIB set urgency = "irrelevant" dan items = [].
Langkah 3: Jika is_disaster_related = true, tentukan urgensi dan ekstrak kebutuhan.

=== ATURAN SUPER KETAT ===
- Kata "lapar"/"laper"/"kelaparan" TANPA konteks bencana = IRRELEVANT (orang biasa juga lapar).
- Kata "sakit"/"pusing"/"demam" TANPA konteks bencana = IRRELEVANT (sakit biasa bukan darurat bencana).
- Kata "tolong" TANPA konteks bencana = IRRELEVANT (bisa saja "tolong belikan makanan").
- Nama orang + keluhan pribadi ("ocha laper", "pandu jahat") = IRRELEVANT.
- Jika RAGU, pilih IRRELEVANT. Lebih baik menolak daripada meloloskan spam.

=== ATURAN URGENSI (hanya jika is_disaster_related = true) ===
- "critical" = nyawa terancam langsung (terjebak, tenggelam, kebakaran aktif, luka parah)
- "high" = kondisi serius (di pengungsian, banjir merendam, kelaparan di posko, lansia/bayi butuh bantuan)
- "medium" = butuh bantuan tapi aman (minta logistik untuk korban bencana)
- "low" = tidak mendesak (laporan informasi situasi bencana saja)

=== ATURAN EKSTRAKSI KEBUTUHAN ===
- WAJIB scan seluruh teks dan temukan SEMUA barang yang disebutkan
- Jika ada angka di dekat nama barang, itu adalah jumlahnya
- Jika tidak ada angka, tulis quantity "sesuai kebutuhan"
- JANGAN pernah return items kosong jika ada penyebutan barang dalam teks!

=== CONTOH ===

Teks: "pandu jahat"
Jawaban: {"is_disaster_related":false,"reason":"Hanya kalimat pribadi/curhat, tidak ada konteks bencana atau keadaan darurat.","urgency":"irrelevant","items":[]}

Teks: "ocha laper belum mandi bauuu"
Jawaban: {"is_disaster_related":false,"reason":"Keluhan sehari-hari tentang lapar dan belum mandi, bukan situasi bencana.","urgency":"irrelevant","items":[]}

Teks: "hari ini panas banget capek"
Jawaban: {"is_disaster_related":false,"reason":"Keluhan cuaca dan kelelahan biasa, bukan bencana alam.","urgency":"irrelevant","items":[]}

Teks: "terminal raja basa penuh dengan penumpang macet total"
Jawaban: {"is_disaster_related":false,"reason":"Kemacetan lalu lintas biasa, bukan bencana alam.","urgency":"irrelevant","items":[]}

Teks: "kami terjebak banjir di rumah sudah 2 hari kelaparan butuh makanan dan air"
Jawaban: {"is_disaster_related":true,"reason":"Korban banjir terjebak dan kelaparan, ini situasi darurat bencana.","urgency":"critical","items":[{"name":"makanan","quantity":"sesuai kebutuhan"},{"name":"air","quantity":"sesuai kebutuhan"}]}

Teks: "saya di posko pengungsian butuh 10 air mineral dan 5 selimut"
Jawaban: {"is_disaster_related":true,"reason":"Permintaan logistik dari posko pengungsian bencana.","urgency":"high","items":[{"name":"air mineral","quantity":"10"},{"name":"selimut","quantity":"5"}]}

Teks: "terjebak longsor butuh evakuasi"
Jawaban: {"is_disaster_related":true,"reason":"Korban terjebak longsor, nyawa terancam.","urgency":"critical","items":[]}

=== TEKS DARI PENGGUNA ===
%s

Jawab HANYA JSON! Format: {"is_disaster_related":bool,"reason":"...","urgency":"...","items":[...]}`, combinedText)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		fmt.Printf("[NLP] Gemini API Error: %v, using regex fallback\n", err)
		finalUrgency := keywordUrgency
		if finalUrgency == "" {
			finalUrgency = "medium"
		}
		itemsJSON, _ := json.Marshal(regexItems)
		return finalUrgency, string(itemsJSON), nil
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		fmt.Println("[NLP] Empty response from Gemini, using regex fallback")
		finalUrgency := keywordUrgency
		if finalUrgency == "" {
			finalUrgency = "medium"
		}
		itemsJSON, _ := json.Marshal(regexItems)
		return finalUrgency, string(itemsJSON), nil
	}

	part := resp.Candidates[0].Content.Parts[0]
	textResp, ok := part.(genai.Text)
	if !ok {
		fmt.Println("[NLP] Invalid response type from Gemini, using regex fallback")
		finalUrgency := keywordUrgency
		if finalUrgency == "" {
			finalUrgency = "medium"
		}
		itemsJSON, _ := json.Marshal(regexItems)
		return finalUrgency, string(itemsJSON), nil
	}

	jsonStr := string(textResp)
	fmt.Printf("[NLP] Raw Gemini response: %s\n", jsonStr)

	// Clean up markdown if any
	jsonStr = strings.TrimSpace(jsonStr)
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	// Parse to validate JSON and extract fields
	var result NLPResult
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		fmt.Printf("[NLP] JSON parse error: %v | raw: %s, using regex fallback\n", err, jsonStr)
		finalUrgency := keywordUrgency
		if finalUrgency == "" {
			finalUrgency = "medium"
		}
		itemsJSON, _ := json.Marshal(regexItems)
		return finalUrgency, string(itemsJSON), nil
	}

	// Normalize urgency
	result.Urgency = strings.ToLower(strings.TrimSpace(result.Urgency))
	validUrgency := false
	for _, u := range []string{"irrelevant", "low", "medium", "high", "critical"} {
		if result.Urgency == u {
			validUrgency = true
			break
		}
	}
	if !validUrgency {
		fmt.Printf("[NLP] Invalid urgency from AI: %q, falling back\n", result.Urgency)
		result.Urgency = "medium"
	}

	// Check is_disaster_related field from AI
	if !result.IsDisasterRelated {
		fmt.Printf("[NLP] AI says NOT disaster-related. Reason: %s\n", result.Reason)
		return "irrelevant", "[]", nil
	}

	if result.Urgency == "irrelevant" {
		fmt.Println("[NLP] AI detected irrelevant report, rejecting.")
		return "irrelevant", "[]", nil
	}

	fmt.Printf("[NLP] AI urgency: %q, Keyword urgency: %q\n", result.Urgency, keywordUrgency)

	// Step 4: Use the HIGHER urgency between AI and keyword detection
	finalUrgency := higherUrgency(result.Urgency, keywordUrgency)
	fmt.Printf("[NLP] Final urgency: %q\n", finalUrgency)

	// Step 5: If AI returned empty items but regex found some, use regex items
	finalItems := result.Items
	if len(finalItems) == 0 && len(regexItems) > 0 {
		fmt.Printf("[NLP] AI returned 0 items but regex found %d, using regex items\n", len(regexItems))
		finalItems = regexItems
	}

	// Convert items back to JSON string
	itemsJSON, _ := json.Marshal(finalItems)
	fmt.Printf("[NLP] Final extracted items: %s\n", string(itemsJSON))

	return finalUrgency, string(itemsJSON), nil
}

// higherUrgency returns whichever urgency level is more severe
func higherUrgency(a, b string) string {
	order := map[string]int{
		"irrelevant": -1,
		"low":        0,
		"medium":     1,
		"high":       2,
		"critical":   3,
	}

	scoreA := order[a]
	scoreB := order[b]

	if scoreB > scoreA {
		return b
	}
	return a
}
