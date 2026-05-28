package seed

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/2mes4/llull/internal/engine"
)

type BookDocument struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

func LoadTextFiles(dir string) ([]string, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, nil
	}

	var texts []string
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		content := strings.TrimSpace(string(data))
		if len(content) > 0 {
			texts = append(texts, content)
			name := strings.TrimSuffix(entry.Name(), ".txt")
			if len(name) > 60 {
				name = name[:60]
			}
			names = append(names, name)
		}
	}

	return texts, names, nil
}

func EmbedFallbackTexts() ([]string, []string) {
	return []string{
		strings.Repeat(`Ramon Llull (1232–1316) va ser un filòsof, poeta, teòleg i místic del Regne de Mallorca. És considerat pioner de la teoria computacional gràcies a la seva Ars Magna, un sistema per generar coneixement mitjançant la combinació mecànica de conceptes. Llull va escriure en català, llatí i àrab. Les seves obres inclouen el Llibre de Meravelles, l'Arbre de la Ciència i el Llibre de l'Ordre de Cavalleria. Va inventar el Cercle Lul·lià, una màquina lògica primitiva que va influir Leibniz i la informàtica moderna. Va viatjar extensament per Europa, el Nord d'Àfrica i l'Orient Mitjà. `, 600),
		strings.Repeat(`L'Ars Magna (Art General) és el nucli del sistema filosòfic de Llull. Utilitza un conjunt de cercles concèntrics amb lletres que representen conceptes fonamentals com bondat, grandesa, eternitat, poder, saviesa, voluntat, virtut, veritat i glòria. En girar aquests cercles, es podien generar totes les combinacions possibles de conceptes, creant una màquina de coneixement universal. Aquest enfocament combinatori va anticipar els algorismes de cerca moderns, la recuperació d'informació i les tecnologies web semàntiques. Leibniz es va inspirar directament en el sistema de Llull. `, 600),
		strings.Repeat(`El Llibre de Meravelles de Llull és una enciclopèdia filosòfica estructurada com una sèrie de diàlegs entre un savi i el seu deixeble. El llibre cobreix temes que van des de la teologia i la filosofia fins a les ciències naturals i l'ètica. Explora les meravelles de la creació a través de la raó, argumentant que el món natural reflecteix atributs divins. El text destaca per la seva prosa clara i l'enfocament sistemàtic a l'organització del coneixement, fent-lo accessible a lectors més enllà dels cercles acadèmics. `, 600),
		strings.Repeat(`L'Arbre de la Ciència és l'obra més completa de Llull. El llibre s'estructura com un arbre amb arrels, tronc, branques, fulles i fruit que representen diferents branques del coneixement. Llull va organitzar tot el coneixement humà en un sistema jeràrquic, començant pels principis fonamentals de l'ésser i progressant pel món natural, la societat humana i la revelació divina. Cada branca de l'arbre correspon a una disciplina o àmbit d'investigació específic. `, 600),
		strings.Repeat(`Llull va fundar una escola a Miramar, a Mallorca, l'any 1276, on tretze frares franciscans van estudiar àrab, filosofia i teologia. L'escola va rebre el patrocini reial del rei Jaume II de Mallorca. La visió educativa de Llull era formar missioners que poguessin debatre racionalment amb erudits musulmans i jueus. El seu enfocament emfatitzava la unitat de la veritat a través de les tradicions religioses, argumentant que la raó podia demostrar els principis fonamentals compartits per totes les fe. Aquest enfocament va influir pensadors com Giordano Bruno i Leibniz. `, 600),
		strings.Repeat(`El Cercle Lul·lià és un dispositiu mecànic format per cercles de paper concèntrics que giren al voltant d'un centre comú. Cada cercle està dividit en seccions que contenen lletres o símbols que representen conceptes fonamentals. En girar els cercles, es poden generar diferents combinacions de conceptes, creant noves proposicions i arguments. Aquest dispositiu és considerat una de les primeres màquines de càlcul i va influir directament en el desenvolupament de la informàtica moderna. `, 600),
	}, []string{"Ramon Llull - Introducció", "Ramon Llull - Ars Magna", "Ramon Llull - Llibre de Meravelles", "Ramon Llull - Arbre de la Ciència", "Ramon Llull - Escola de Miramar", "Ramon Llull - Cercle Lul·lià"}
}

func splitIntoChunks(text string, chunkSize int) []string {
	paragraphs := strings.Split(text, "\n")
	var chunks []string
	var current strings.Builder

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if len(p) > chunkSize {
			if current.Len() > 0 {
				chunks = append(chunks, strings.TrimSpace(current.String()))
				current.Reset()
			}
			for _, seg := range splitLongString(p, chunkSize) {
				chunks = append(chunks, seg)
			}
			continue
		}
		if current.Len()+len(p)+1 > chunkSize && current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(p)
	}
	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	return chunks
}

func splitLongString(s string, maxLen int) []string {
	var parts []string
	runes := []rune(s)
	for i := 0; i < len(runes); i += maxLen {
		end := i + maxLen
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[i:end]))
	}
	return parts
}

func GenerateDocumentsFromTexts(texts []string, names []string, targetCount int) []BookDocument {
	rng := rand.New(rand.NewSource(42))
	docs := make([]BookDocument, 0, targetCount)

	chunkSize := 1500

	for i, text := range texts {
		bookName := names[i]

		chunks := splitIntoChunks(text, chunkSize)

		for j, chunk := range chunks {
			if len(docs) >= targetCount {
				break
			}

			title := fmt.Sprintf("%s - Fragment %d", bookName, j+1)

			weight := polarizeWeight(rng)

			docType := "fragment"
			if j == 0 {
				docType = "introduction"
			}

			docs = append(docs, BookDocument{
				ID: fmt.Sprintf("doc-%04d", len(docs)),
				Fields: map[string]interface{}{
					"title":   title,
					"content": chunk,
					"source":  bookName,
					"type":    docType,
					"weight":  weight,
				},
			})
		}
	}

	if len(docs) > targetCount {
		docs = docs[:targetCount]
	}

	return docs
}

func extractTitle(chunk string, idx int, source string) string {
	lines := strings.Split(chunk, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 10 && len(line) < 120 {
			return line
		}
	}
	return fmt.Sprintf("Fragment %d - %s", idx, source)
}

func polarizeWeight(rng *rand.Rand) float64 {
	v := rng.Float64()
	if v < 0.33 {
		return rng.Float64() * 0.2
	}
	if v < 0.66 {
		return 0.2 + rng.Float64()*0.3
	}
	return 0.7 + rng.Float64()*0.3
}

func GenerateSeedFile(path string, texts []string, names []string, targetCount int) error {
	docs := GenerateDocumentsFromTexts(texts, names, targetCount)
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling seed data: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func LoadSeedFile(path string) ([]BookDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading seed file: %w", err)
	}
	var docs []BookDocument
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parsing seed file: %w", err)
	}
	return docs, nil
}

func ToIndexPayloads(docs []BookDocument) []engine.IndexPayload {
	payloads := make([]engine.IndexPayload, len(docs))
	for i, doc := range docs {
		payloads[i] = engine.IndexPayload{
			ID:        doc.ID,
			Action:    "INDEX",
			Fields:    doc.Fields,
			UpdatedAt: time.Now().Unix(),
		}
	}
	return payloads
}
