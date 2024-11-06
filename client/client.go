package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal("Erro ao criar requisição para o servidor:", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Erro ao fazer requisição para o servidor:", err)
	}
	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Erro ao ler resposta do servidor:", err)
	}

	var response map[string]string
	if err := json.Unmarshal(resBytes, &response); err != nil {
		log.Fatal("Erro ao parsear JSON:", err)
	}

	bid, ok := response["bid"]
	if !ok {
		log.Fatal("Campo 'bid' não encontrado na resposta")
	}

	file, err := os.Create("client/cotacao.txt")
	if err != nil {
		log.Fatal("Erro ao criar arquivo:", err)
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("Dólar: %s\n", bid)); err != nil {
		log.Fatal("Erro ao escrever no arquivo:", err)
	}

	log.Println("Cotação salva no arquivo: cotacao.txt")
}
