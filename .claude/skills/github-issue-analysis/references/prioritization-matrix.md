# Prioritization Matrix

Five dimensions, each scored 1–10, with configurable weights.

## Default Weights

| Dimension | Weight | Rationale |
|---|---|---|
| Demand | 30% | Volume of asks and engagement — what users want most |
| Community Interest | 25% | Breadth of interest — how many different people care |
| Urgency | 20% | How long people have been waiting and recent momentum |
| Feasibility | 15% | Estimated effort — quick wins score higher |
| Strategic Alignment | 10% | Fit with project direction, labels, maintainer signals |

**Total: 100%**

The user can override these weights. Present the defaults and ask before
scoring.

## Adaptive Scoring

Scoring thresholds must scale with dataset size to produce meaningful
differentiation. A theme with 11 issues out of 15 total is dominant; a theme
with 11 issues out of 500 is modest. Use the approach below.

### How it works

Before scoring, compute the **theme median** and **theme max** for each
quantitative signal across all discovered themes. Then score each theme
relative to its peers using a percentile-based 1–10 scale:

```
score = 1 + 9 × (theme_value - theme_min) / (theme_max - theme_min)
```

Round to one decimal. If all themes have the same value for a signal, score
everyone 5 (neutral). This ensures the full 1–10 range is used regardless of
whether the repo has 15 or 5,000 issues.

Apply this relative scoring to all quantitative signals in Demand, Community
Interest, and Urgency. Feasibility and Strategic Alignment remain
judgment-based (absolute scales).

### Fallback for small repos (≤30 issues)

For repos with 30 or fewer total issues, the absolute thresholds below
provide reasonable differentiation. Use them instead of relative scoring.

## Dimension Scoring — Absolute Thresholds (small repos)

### Demand (weight: 30%)

Measures the raw volume of signal for a theme.

| Signal | Scoring |
|---|---|
| Issue count in theme | 1 issue = 2, 2–3 = 4, 4–6 = 6, 7–10 = 8, 11+ = 10 |
| Total reactions (thumbs-up, heart, rocket) | 0 = 1, 1–5 = 3, 6–15 = 5, 16–30 = 7, 31+ = 9 |
| Total comments across theme | 0–2 = 1, 3–10 = 3, 11–25 = 5, 26–50 = 7, 51+ = 9 |

**Theme demand score** = average of the three signals above, capped at 10.

### Community Interest (weight: 25%)

Measures breadth — how many different people are asking, not just how loud
one person is.

| Signal | Scoring |
|---|---|
| Unique authors across theme issues | 1 = 2, 2–3 = 4, 4–6 = 6, 7–10 = 8, 11+ = 10 |
| Unique commenters across theme | 0–1 = 1, 2–5 = 3, 6–10 = 5, 11–20 = 7, 21+ = 9 |
| Cross-reference mentions (issues linking to each other) | Bonus +1 if present |

**Theme community score** = average of signals, capped at 10.

### Urgency (weight: 20%)

Measures how long the community has been waiting and whether momentum is
building.

| Signal | Scoring |
|---|---|
| Age of oldest issue in theme | <30d = 2, 30–90d = 4, 90d–6m = 6, 6m–1y = 8, >1y = 10 |
| Issues created in last 90 days | 0 = 2, 1 = 4, 2–3 = 6, 4–5 = 8, 6+ = 10 |
| Recent comment activity (last 30d) | None = 1, 1–3 comments = 4, 4–10 = 7, 11+ = 10 |

**Theme urgency score** = average of signals, capped at 10.

## Dimension Scoring — Relative Thresholds (large repos)

For repos with more than 30 issues, apply the relative scoring formula to
these signals per theme:

**Demand:**
- Issue count in theme (relative to other themes)
- Total reactions across theme (relative)
- Total comments across theme (relative)

**Community Interest:**
- Unique authors across theme issues (relative)
- Unique commenters across theme (relative)
- Cross-reference bonus: +0.5 if issues within the theme link to each other

**Urgency:**
- Age of oldest issue in theme (relative — older = higher)
- Issues created in last 90 days (relative — more recent = higher)
- Recent comment activity in last 30 days (relative)

Each dimension score = average of its signal scores, capped at 10.

### Feasibility (weight: 15%)

Estimates implementation complexity. Higher score = easier to implement.
This is necessarily a rough estimate based on issue content. Uses absolute
scoring regardless of repo size.

| Complexity Indicator | Score |
|---|---|
| Simple config/docs change | 9–10 |
| Single resource or data source addition with clear API | 7–8 |
| Multiple related resources or moderate API complexity | 5–6 |
| Cross-cutting concern (auth, framework change) | 3–4 |
| Major architectural change or unclear scope | 1–2 |

Assess based on issue descriptions, required API surface, and similar work
in comparable providers. When uncertain, score 5 (neutral).

### Strategic Alignment (weight: 10%)

Measures fit with project direction. This dimension is subjective and
benefits most from user input. Uses absolute scoring regardless of repo size.

| Signal | Score |
|---|---|
| Maintainer has commented positively or assigned | 8–10 |
| Labels indicate planned/accepted (e.g. "enhancement", "accepted") | 6–7 |
| Aligns with stated project goals or upstream direction | 5–7 |
| Neutral — no signal either way | 5 |
| Conflicts with project direction or marked "wontfix" | 1–3 |

If the user states strategic priorities (e.g. "enterprise readiness",
"upstream parity"), boost themes that align.

## Computing the Final Score

```
priority_score = (demand × 0.30) + (community × 0.25) + (urgency × 0.20)
               + (feasibility × 0.15) + (alignment × 0.10)
```

Round to one decimal place. Rank themes by descending score.

## Recommended Actions

Based on the priority score and feasibility, assign a recommended action:

| Score Range | Feasibility | Action |
|---|---|---|
| 7.0+ | High (7+) | **Quick Win** — implement soon |
| 7.0+ | Low (<5) | **Needs Design** — high demand but complex, invest in design |
| 5.0–6.9 | Any | **Plan** — schedule for upcoming cycle |
| 3.0–4.9 | Any | **Defer** — revisit next planning cycle |
| <3.0 | Any | **Backlog** — low signal, park for now |

## Adjusting Weights

The user may want to shift the balance. Common adjustments:

- **"We need quick wins"** → Increase Feasibility to 30%, reduce Urgency to 10%
- **"Strategic alignment matters most"** → Increase Alignment to 25%, reduce Demand to 20%
- **"Community-driven project"** → Increase Community Interest to 35%, reduce Alignment to 5%
- **"We have a backlog crisis"** → Increase Urgency to 30%, reduce Feasibility to 10%

Present the adjusted weights and regenerate scores when the user requests changes.
