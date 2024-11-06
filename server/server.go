package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	http.HandleFunc("/cotacao", handler)
	log.Println("Server rodando na porta :8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		http.Error(w, "Erro ao criar requisição para API", http.StatusInternalServerError)
		log.Println("Erro ao criar requisição:", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Erro ao fazer requisição para API", http.StatusInternalServerError)
		log.Println("Erro ao fazer requisição:", err)
		return
	}
	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Erro ao ler resposta da API", http.StatusInternalServerError)
		log.Println("Erro ao ler resposta:", err)
		return
	}

	response := make(map[string]map[string]interface{})
	if err := json.Unmarshal(resBytes, &response); err != nil {
		http.Error(w, "Erro ao parsear JSON", http.StatusInternalServerError)
		log.Println("Erro ao parsear JSON:", err)
		return
	}
	bid, ok := response["USDBRL"]["bid"].(string)
	if !ok {
		http.Error(w, "Campo 'bid' não encontrado", http.StatusInternalServerError)
		log.Println("Campo 'bid' não encontrado")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"bid": bid})

	const dbFile = "cotacoes.db"
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Println("Erro ao abrir banco de dados:", err)
		return
	}
	defer db.Close()

	const createTable = `
		CREATE TABLE IF NOT EXISTS cotacao (
			id INTEGER NOT NULL PRIMARY KEY,
			data DATETIME NOT NULL,
			registro TEXT NOT NULL
		);`

	if _, err := db.Exec(createTable); err != nil {
		log.Println("Erro ao criar tabela:", err)
		return
	}

	insertCtx, insertCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer insertCancel()

	const insert = `INSERT INTO cotacao (data, registro) VALUES (?, ?)`
	if _, err := db.ExecContext(insertCtx, insert, time.Now(), bid); err != nil {
		log.Println("Erro ao inserir dados no banco:", err)
		return
	}

	log.Println("Cotação registrada no banco de dados:", bid)
}
