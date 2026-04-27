# 🧠 AI Newsletter via Telegram (Go)

## 🎯 Objetivo

Criar um sistema que:

* Coleta notícias de múltiplas fontes (RSS/APIs)
* Deduplica e prepara itens para o digest
* Gera resumos usando IA
* Envia diariamente via Telegram

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

### Fase 5 — Personalização

* [ ] Perfis (você / namorada)
* [ ] **Interesses no Telegram**: até **5** temas por perfil (ex.: AI, tech, futebol, NBA, bitcoin) — ver *Interesses: lista pronta vs texto livre* abaixo
* [ ] **Cooldown de alteração de temas**: após confirmar os temas, o utilizador **só pode voltar a mudar os temas após 24 horas desde essa confirmação** (regra de produto; implementar com `confirmed_at` no perfil ou equivalente)
* [ ] Mapear cada interesse → lista de **keywords / sinônimos** (prompt da IA e relevância)
* [ ] **No onboarding** (quando o utilizador confirma os temas): chamada na API de IA com **RAG de catálogo de fontes** para sugerir **3 melhores feeds RSS por tema** — **não** repetir este passo em cada execução do digest diário
* [ ] **Persistir** as URLs dos feeds aprovados no perfil; o pipeline diário **só lê estas URLs guardadas** (variáveis de ambiente globais ficam para MVP, defaults ou deploy sem perfis)
* [ ] Validar operacionalidade de cada feed sugerido (`HTTP 200`, parse RSS ok, itens recentes) antes de salvar no perfil
* [ ] **Pesos dos tiers** por perfil ou global (override dos defaults do `domain`)
* [ ] Filtragem personalizada

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

**Personalização (Fase 5):** cada **perfil** (ou ambiente) pode **sobrescrever** esses inteiros na config quando existir multi‑audiência.

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

### Interesses: lista pronta vs texto livre (Telegram, máx. 5)

**O que é mais usado na indústria:** fluxo principal com **lista pré-curada** (multi‑select, botões inline ou menu com busca) + **limite fixo** (aqui: 5). Motivos: ortografia consistente, i18n, keywords estáveis por trás, menos abuso, onboarding rápido.

**Texto livre** (“escreva seus 5 temas”) aparece em produtos estilo *Google Alerts* ou power users: mais flexível, mas exige **normalização** (sinónimos, stemming, ou um passo de IA para mapear frase → tags internas) e gera mais ruído na relevância.

**Híbrido (recomendado para o Neurofeed):** **catálogo grande pesquisável** (categorias: tecnologia, desporto, mercados…) para a maior parte dos utilizadores + **0–1 slots opcionais de texto curto** (“outro: ___”) com validação (tamanho máx., lista negra, mapeamento manual ou IA leve para tag interna). Assim cobres *NBA* e *bitcoin* sem depender de o utilizador escrever bem *Ethereum* vs *etherium*.

**Resumo:** começa com **lista pré-feita + busca**; acrescenta **texto livre limitado** só se métricas pedirem.

---

### Descoberta de fontes por tema (IA + RAG)

**Quando corre:** **Apenas no onboarding** — depois de o utilizador confirmar até **5 temas** (e respeitando o **cooldown de 24 horas desde a última confirmação** antes de poder mudar temas outra vez; ver Fase 5). O digest diário **não** invoca LLM+RAG para escolher URLs; usa as feeds **já guardadas** no perfil.

**Depois do onboarding:** o utilizador **mantém** as URLs validadas enquanto não mudar os temas; **nova mudança de temas** só é permitida quando tiverem passado **24 horas** desde o `confirmed_at` da última confirmação (ou via futura funcionalidade explícita “atualizar fontes”).

Quando o utilizador finalizar a seleção de temas, o sistema executa um passo de descoberta assistida:

1. Para cada tema, faz **1 chamada na API de IA** com contexto RAG (catálogo interno de fontes já conhecidas e seus metadados: idioma, categoria, país, qualidade histórica e URL do feed).
2. A IA devolve **top 3 fontes RSS** para cada tema, com justificativa curta.
3. Antes de persistir, o backend valida cada feed candidato:
   - responde com `HTTP 200`;
   - parseia como RSS/Atom com sucesso;
   - possui itens recentes (janela configurável, ex.: últimos 7 dias).
4. Só feeds aprovados entram no perfil; falhas voltam para fallback (fontes default por tema) e log de observabilidade.

Objetivo: garantir que a personalização não depende apenas de "nome bonito de fonte", mas de feeds realmente operacionais para o pipeline diário — e que o custo e a variabilidade do LLM+RAG ficam **no momento de configuração do perfil**, não em cada corrida agendada.

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
