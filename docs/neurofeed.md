# 🧠 AI Newsletter via Telegram (Go)

## 🎯 Objetivo

Criar um sistema que:

* Coleta notícias de múltiplas fontes (RSS/APIs)
* Deduplica e prepara itens para o digest
* Gera resumos usando IA
* Envia diariamente via Telegram (**só entrega**: assuntos e fontes vêm de config fixa; o utilizador **não** configura temas no chat)

---

## 🏗️ Arquitetura Geral

```
[cron]
   ↓
coleta (RSS/APIs)
   ↓
deduplicação / preparação
   ↓
IA (resumo)
   ↓
formatação
   ↓
Telegram Bot
```

---

## 🛠️ Stack

* Linguagem: Go
* Scheduler: cron (ou gocron)
* HTTP client: net/http
* Parsing RSS: github.com/mmcdole/gofeed
* IA: OpenAI API
* Output: Telegram Bot API

---

## 🗺️ Roadmap

### Fase 1 — MVP (1–2 dias)

* [x] Criar bot no Telegram
* [x] Implementar envio simples de mensagem
* [x] Ler 1 feed RSS
* [x] Extrair título + link
* [x] Enviar para Telegram

### Fase 2 — Múltiplas fontes

* [x] Adicionar múltiplos RSS feeds
* [x] Cada feed na config com **tier** (`primary`, `expert`, `news`, `community`); ingest preenche `Article.SourceTier`
* [x] Normalizar estrutura dos dados
* [x] Deduplicação básica (por título)

### Fase 3 — Integração com IA (3.1 → 3.2 → 3.3)

* [ ] **3.1** — Cliente HTTP OpenAI (ou compatível), env (`LLM_API_KEY`, modelo, timeouts); chamada mínima com `context` para validar I/O
* [ ] **3.2** — Prompts do digest conforme spec; lote de artigos → pedido; saída estruturada / parseável para o pipeline
* [ ] **3.3** — Limites de artigos/tokens, heurísticas de tamanho e clareza; `Summarizer` real na pipeline + testes (`httptest`)

### Fase 4 — UX da mensagem

* [ ] Separar por categorias
* [ ] Adicionar emojis
* [ ] Formatação Markdown
* [ ] Links clicáveis

### Fase 5 — Destinatários e temas fixos (sem escolha no Telegram)

* [ ] **Telegram só recebe:** o utilizador **não** escolhe temas nem interage com fluxos de configuração no chat; só recebe mensagens já formatadas (Fase 4+) com **secções por assunto** definidos pelo operador.
* [ ] **Perfis / destinatários** (ex.: várias pessoas ou `chat_id` distintos): cada um tem **conjunto de temas e feeds fixos na config** (código ou ficheiro versionado) — não há onboarding nem confirmação pelo utilizador final.
* [ ] Mapear cada **tema atribuído** ao destinatário → lista de **keywords / sinônimos** (prompt da IA e relevância), quando a filtragem por tema fizer sentido
* [ ] **Fontes por tema = lista fixa curada** no produto: cada tema mapeia para **URLs RSS** (com tier por URL). **Sem** LLM nem RAG para descobrir feeds.
* [ ] O pipeline diário resolve **só a partir da config** as URLs por destinatário (união das listas dos seus temas, dedupe por URL); variáveis de ambiente globais (`NEUROFEED_RSS_FEEDS`, etc.) podem continuar como MVP para um único destinatário
* [ ] Validar operacionalidade dos feeds em job de manutenção ou arranque (`HTTP 200`, parse RSS ok, itens recentes onde aplicável); feeds inválidos: log + exclusão temporária ou fallback definido pelo operador
* [ ] **Pesos dos tiers** por destinatário ou global (override dos defaults do `domain`)
* [ ] Filtragem por tema / destinatário quando aplicável

### Fase 6 — Robustez

* [ ] Retry de falhas
* [ ] Logs
* [ ] Timeout em requests
* [ ] Cache simples

---

## 📰 Sistema de Coleta (RSS)

### Estrutura de dados

```go
type Article struct {
    Title       string
    Link        string
    Description string
    Source      string
    SourceTier  SourceTier // ver Camadas de fontes (tiers)
    Published   time.Time
}
```

---

### Leitura de RSS

```go
fp := gofeed.NewParser()
feed, _ := fp.ParseURL(url)

for _, item := range feed.Items {
    article := Article{
        Title: item.Title,
        Link: item.Link,
        Description: item.Description,
        SourceTier: tierDaConfigParaEsteFeedURL,
        Published: *item.PublishedParsed,
    }
}
```

---

### Camadas de fontes (tiers)

Política editorial do Neurofeed: **cada URL de feed na config** recebe um **tier**. O ingest copia isso para `Article.SourceTier` (ver tabela). Em `domain`, `SourceTier.ScoreWeight()` e constantes como `DefaultTierWeightPrimary` guardam **pesos editoriais padrão** para ranking ou personalização futuros — hoje o binário só usa o tier na ingestão e metadados do artigo.

| Tier | O que é | Exemplos de RSS / uso |
|------|---------|------------------------|
| **primary** | Fatos ou artefatos oficiais; rastreável até a origem. | Blogs oficiais de produto (release notes), feeds de órgãos reguladores, repositórios com `releases.atom`, datasets publicados como feed. |
| **expert** | Síntese técnica ou acadêmica estável. | Revistas / periódicos com resumo, blogs de especialistas com referências, documentação de consórcios (W3C, IETF) quando exposta como feed. |
| **news** | “O que mudou hoje” com redação. | Agências e veículos que linkam primários, newsletters jornalísticas em RSS. |
| **community** | Opinião, hype, fio social republicado. | Substack opinativo sem obrigação de primário, agregadores fracos, comentário puro. |

**Pesos padrão** (constantes em `domain`, ex.: `DefaultTierWeightPrimary`): `primary` +4, `expert` +3, `news` +2, `community` −1, não configurado 0.

**Multi-destinatário (Fase 5):** cada **perfil** na config pode **sobrescrever** esses inteiros quando houver vários destinatários com políticas diferentes.

**Regra de ouro:** tiers alinham o digest à hierarquia de fontes; **não substituem** sozinhos o resumo com IA nem o contexto editorial que vêm da Fase 3 em diante.

---

### Deduplicação

```go
map[string]bool // chave: título normalizado
```

* lower case
* remove pontuação
* opcional: hash

---

### Temas e destinatários (só configuração, não Telegram)

**Modelo:** quais **assuntos / secções** aparecem no digest e que **feeds** alimentam cada assunto são definidos **só pelo operador** (config versionada ou código). O utilizador no Telegram **não** escolhe temas, não confirma listas e não há cooldown de “mudança de interesses” no produto — alterações são **deploy / editar config**.

**Formato:** na Fase 4+, o texto enviado ao Telegram pode agrupar itens por assunto (cabeçalhos por tema) para leitura clara; esses rótulos vêm da mesma config que fixa os RSS por tema.

---

### Fontes por tema (lista fixa curada)

**Política:** cada **tema** usado no produto tem uma **lista fixa de feeds RSS** (URLs + tiers) — p.ex. struct em Go, JSON em repo, ou tabela estática. **Não** há descoberta de feeds por LLM, RAG, nem busca na internet no fluxo do Neurofeed.

**Resolução:** para cada destinatário, o backend calcula o conjunto de URLs a partir dos **temas atribuídos a esse destinatário na config** (união das listas fixas, dedupe por URL), com validação opcional em arranque ou job (`HTTP 200`, parse RSS/Atom ok).

**Digest diário:** usa **apenas** URLs derivadas dessa config (ou env global no MVP de um único `chat_id`). A IA entra **só** na fase de resumo / digest (Fase 3+), não na escolha de fontes.

**Manutenção:** quando um feed fixo deixa de responder, o operador atualiza a lista curada (release) ou um job marca o feed como indisponível até correção — sem passar por modelo para “encontrar substituto”.

---

## 🤖 Prompt de IA (Resumo de Qualidade)

### Objetivo

Evitar:

* resumos genéricos
* linguagem vaga
* falta de contexto

---

### Prompt Base

```
Você é um analista de notícias altamente objetivo.

Sua tarefa é resumir as notícias abaixo de forma útil e direta.

Regras:
- Não seja genérico
- Extraia o ponto principal de cada notícia
- Foque no impacto (por que isso importa?)
- Evite frases vagas como "isso é importante"
- Use linguagem simples

Formato de saída:

🧠 Resumo do dia

Para cada notícia:
- Título reescrito (curto e claro)
- 1-2 frases explicando o que aconteceu
- 1 frase explicando o impacto

Dados:
{{ARTIGOS}}
```

---

### Prompt Avançado (melhor qualidade)

```
Atue como um analista de mercado e tecnologia.

Para cada notícia:
1. Explique o fato principal
2. Explique o contexto (se necessário)
3. Explique o impacto prático

Se a notícia não for relevante, ignore.

Seja conciso, mas informativo.

Saída:
- Bullet points
- Máximo 8 notícias
- Linguagem clara e direta

Evite:
- redundância
- opinião sem base
- jargão desnecessário

Dados:
{{ARTIGOS}}
```

---

## ✨ Formatação Final (Telegram)

Exemplo:

```
🧠 *Resumo do dia*

📈 *Mercado*
- Bitcoin sobe 3% após...
  Impacto: movimento pode indicar...

🤖 *Tecnologia*
- Empresa X lança...
  Impacto: isso afeta...

🔗 Leia mais
```

---

## 🚀 Evoluções Futuras

* Ranking por relevância com IA
* Clusterização de notícias
* Detecção de tendências
* Áudio (TTS)
* Interface web simples

---

## 🧠 Princípio-chave

> O valor não está em enviar notícias —
> está em filtrar e resumir melhor que qualquer feed comum.

---

## ✅ Definição de sucesso

* Você realmente lê todo dia
* Não sente que perdeu tempo
* Descobre coisas relevantes rapidamente

---

Se isso acontecer, o sistema está funcionando.
