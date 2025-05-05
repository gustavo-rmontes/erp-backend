# 🧠 ERP Inteligente com Tutor de IA por Voz  
**On Smart Tech – Plataforma Corporativa com Automação e IA Integrada**

---

## 🚀 Visão Geral

Este projeto é uma plataforma ERP modular, executada localmente (on-premises), com integração nativa a agentes de inteligência artificial, incluindo um **Tutor de IA com interação por voz** que guia os usuários em tempo real.

Criado pela **On Smart Tech**, o sistema é desenvolvido para uso interno e também para comercialização como solução B2B/B2C.

---

## 🔍 Principais Funcionalidades

- ⚙️ Módulos iniciais: **Financeiro, Vendas e Marketing**
- 🧠 Agentes de IA integrados: previsões, automações e insights
- 🗣️ Tutor de IA por voz: explicação de telas, campos e fluxos
- 🔐 Controle de acesso por perfil (RBAC) com autenticação JWT
- 🧱 Arquitetura moderna com **Go (back-end)** e **Python (IA)**
- 🖥️ Front-end SPA com React ou Vue
- 📦 Execução local via Docker + Docker Compose

---

## 🧠 Stack Tecnológico

| Camada         | Tecnologia                            |
|----------------|----------------------------------------|
| Back-end ERP   | Go, Gin, Viper, JWT, PostgreSQL        |
| IA e Tutor     | Python, FastAPI, Whisper, Coqui TTS, LangChain |
| Front-end      | React ou Vue, Vite/Webpack, Pinia/Redux |
| Comunicação    | REST API, RabbitMQ/NATS                |
| Deploy         | Docker, Docker Compose, GitHub Actions |

---

## 📁 Estrutura do Projeto

```bash
erp-inteligente/
├── backend/       # ERP core em Go (Clean Architecture)
├── ai/            # Agentes e Tutor IA (Python)
├── frontend/      # Interface do usuário (SPA)
├── docker/        # Dockerfiles e orquestração
├── docs/          # Documentação, diagramas, prints
├── scripts/       # Migração, backup, setup
├── .env.example   # Variáveis de ambiente de exemplo
├── .gitignore     # Arquivos ignorados pelo Git
├── .dockerignore  # Arquivos ignorados pelo Docker
├── LICENSE        # Licença Apache 2.0
├── README.md      # Este arquivo
```
---

## ⚙️ Como Rodar Localmente
🔧 Pré-requisitos
Docker e Docker Compose instalados

Git instalado

Portas 8080 (API), 5001 (IA), 5173 (Front-end) livres

---

## 🗣️ Como funciona o Tutor de IA por Voz?
Reconhece comandos de voz (via Whisper ou Vosk)

Identifica elementos da tela (DOM, contexto visual)

Responde com áudio (via Coqui TTS ou ElevenLabs)

Permite ativação manual com segurança e privacidade

Exemplo de comando:
“Tutor, como faço um novo pedido de venda?”

📸 Prints do Sistema (em breve)
📷 Em breve adicionaremos imagens da interface:

docs/screenshots/dashboard.png

---

## 👥 Perfis e Permissões (RBAC)
- Perfil:	Acesso a Módulos
- admin:	Todos os módulos
- finance_user:	Financeiro
- marketing_user:	Marketing
- sales_user:	Vendas

---  

🧪 Testes

Go: `go test ./...`

Python: `pytest`

Front-end: `npm run test`

---

## 🗺️ Roadmap
Veja o roadmap completo em docs/roadmap.md

- Fase	Status
- Estrutura do projeto	✅ Concluído
- Integração de módulos ERP	🔄 Em desenvolvimento
- Tutor IA por voz	🔄 MVP inicial
- Interface com IA (Agentes)	🔄 Em progresso
- Deploy e monitoramento	🔲 Planejado

---


## 📄 Licença
Este projeto está licenciado sob a Apache License 2.0.
Veja o arquivo LICENSE para mais detalhes.

---

## 👨‍💻 Contato
Desenvolvido por On Smart Tech
📧 atendimento@onsmart.com.br
🌐 https://www.onsmart.com.br 

    


