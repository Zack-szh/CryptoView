# CryptoView

Useful commands: 
- start docker: 
    docker-compose up 
- close docker while deleting database: 
    docker-compose down -v 
- launch interactive psql shell: 
    docker exec -it cryptoview-postgres-1 psql -U cryptoview -d cryptoview

- launch streaming service and add data to DB: 
    go run ./services/market-data/cmd/stream/main.go

- launch api service and query data from DB: 
    go run ./services/api/cmd/api/main.go

- launch frontend on http://localhost:3000/: 
    cd frontend
    npm run dev

- launch agent server on http://localhost:8000:
    cd ./services/ml && uv run uvicorn cryptoview_ml.server:app --port 8000 --reload

- serve local model on http://localhost:8081
    mlx_lm.server --port 8081 --model Qwen3.6-35B-A3B-MLX-4bit