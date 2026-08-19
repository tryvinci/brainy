# S0 WRITE_MISS hand-adjudication (product lane, 2026-08-19)

**Not a product change.** 50 of 120 product-lane `WRITE_MISS` rows from
`locomo-s0-20260819-product-recall-s1-828403`, sampled with seed 19.
Labels use the P3 buckets. Gold check on this ledger was **lexical substring**;
many rows have 800–1,200 compiled facts, so a substring miss is not proof the
claim was never written.

## Histogram (n=50)

| Bucket | n | Notes |
| --- | ---: | --- |
| TRUE_WRITE_MISS | 32 | Generated answer unrelated to the gold object |
| TEMPORAL_LOSS | 7 | Date/duration missing or wrong while the event exists nearby |
| PARAPHRASE | 4 | Related claim in different words; lexical gold failed |
| ENTITY_LINK | 3 | Named entities (pets, people) not bound |
| WRONG_VALUE | 2 | Polarity or count off (yes/no, two times) |
| ORACLE_ERROR | 1 | Gold date present in the answer; punctuation mismatch |
| RELATION_PROJECTION | 1 | Gift/recipient list not joined |

Retrieval was **not** assessed before this label on the original run, so some
TRUE_WRITE_MISS rows may still be RETRIEVAL_MISS after P3 reorder. This sample
is evidence that WRITE_MISS=120 is a mixed bucket, not a single compiler gap.

## Rows

| ID | Group | Bucket | Gold (short) | Why |
| --- | --- | --- | --- | --- |
| conv-48-q116 | SH | TRUE_WRITE_MISS | connected to her body | Unrelated book dump |
| conv-26-q60 | MH | TRUE_WRITE_MISS | clarinet and violin | Pottery/race, no instruments |
| conv-49-q37 | MH | TRUE_WRITE_MISS | salad, grilled salmon… | Health slogan, no meals |
| conv-50-q112 | SH | TRUE_WRITE_MISS | satisfying and worth the hard work | Card-night / sunlight |
| conv-47-q40 | MH | TRUE_WRITE_MISS | swimming, catching frisbees… | "Opening Moves" |
| conv-30-q0 | temporal | ORACLE_ERROR | 19 January, 2023 | Answer contains `19 January 2023` |
| conv-47-q32 | temporal | TRUE_WRITE_MISS | Toronto, Canada | Dog outing, no city |
| conv-41-q86 | SH | TRUE_WRITE_MISS | His car broke down | `not in memory` |
| conv-43-q53 | OD | TRUE_WRITE_MISS | Sprinting, long-distance, boxing | "German" |
| conv-43-q3 | OD | TRUE_WRITE_MISS | C. S.Lewis | `not in memory` |
| conv-47-q46 | temporal | TEMPORAL_LOSS | August 27, 2022 | Apartment/bar, no date |
| conv-42-q116 | SH | PARAPHRASE | watching fantasy and sci-fi movies | Nintendo/games nearby |
| conv-47-q141 | SH | TRUE_WRITE_MISS | Fortnite, Overwatch, Apex | Gamepad, no titles |
| conv-30-q57 | SH | PARAPHRASE | build relationships… stay positive | Motivational dump, weak overlap |
| conv-48-q16 | SH | ENTITY_LINK | Susie, Seraphim | Walking Dead / student |
| conv-42-q56 | MH | TRUE_WRITE_MISS | Turtles. | Dedication filler |
| conv-26-q126 | SH | TRUE_WRITE_MISS | Horseback riding | Biking/volunteering |
| conv-49-q78 | MH | RELATION_PROJECTION | To Sam, to his friends… | `not in memory` |
| conv-43-q79 | SH | TRUE_WRITE_MISS | Nike / Gatorade deals | Pro-ball movie dump |
| conv-42-q143 | SH | TRUE_WRITE_MISS | Touched | Vague gold, unrelated |
| conv-42-q84 | OD | TRUE_WRITE_MISS | No; both faced setbacks | "submitting" |
| conv-49-q71 | OD | TRUE_WRITE_MISS | Christmas | Cousin wedding photo |
| conv-50-q125 | SH | TRUE_WRITE_MISS | at a festival | Nature/sunlight |
| conv-42-q123 | SH | TRUE_WRITE_MISS | stuffed animal | turtles/kitchen |
| conv-26-q36 | temporal | TEMPORAL_LOSS | weekend before 17 July 2023 | Mentorship journey, no date |
| conv-47-q102 | SH | TRUE_WRITE_MISS | Reading under the covers | Social overlap, no reading |
| conv-48-q44 | temporal | TEMPORAL_LOSS | 9 April, 2023 | Different February date |
| conv-50-q101 | SH | TRUE_WRITE_MISS | rap-industry podcast | "Friend" |
| conv-26-q81 | OD | WRONG_VALUE | No; adopting children | Answer `yes` |
| conv-48-q155 | SH | TRUE_WRITE_MISS | yoga, meditation, walks | Book dump |
| conv-26-q135 | SH | PARAPHRASE | hurt, break from pottery | Pottery present, injury not |
| conv-47-q53 | temporal | TEMPORAL_LOSS | six months | Pride/game, no duration |
| conv-44-q38 | temporal | TEMPORAL_LOSS | weekend before Oct 24, 2023 | Food outing, no date |
| conv-43-q49 | MH | WRONG_VALUE | two times | Motivational filler |
| conv-48-q145 | SH | ENTITY_LINK | Video games and Susie | Walks/beach, no pet |
| conv-26-q90 | SH | TEMPORAL_LOSS | married for 5 years | Husband present, duration gone |
| conv-43-q99 | SH | TRUE_WRITE_MISS | Prepping the feast | Rothfuss / movie theme |
| conv-48-q177 | SH | TRUE_WRITE_MISS | new level of joy | Exams/deadlines |
| conv-48-q98 | SH | TEMPORAL_LOSS | when she was 10 | Goals filler |
| conv-49-q99 | SH | TRUE_WRITE_MISS | dream interpretation book | Check-up / health |
| conv-43-q102 | SH | TRUE_WRITE_MISS | culture, food | Nature connection |
| conv-50-q81 | SH | TRUE_WRITE_MISS | music gear and microphone | "my job, loving" |
| conv-43-q151 | SH | TRUE_WRITE_MISS | Academic and sports successes | "Hey Tim" |
| conv-42-q117 | SH | PARAPHRASE | strawberry | Ice cream present, flavor not |
| conv-47-q104 | SH | TRUE_WRITE_MISS | Vancouver | Career-skills dump |
| conv-47-q22 | MH | TRUE_WRITE_MISS | animal shelter, homeless, hospital | D&D title |
| conv-41-q7 | MH | TRUE_WRITE_MISS | A doll, a film camera | "Exploring" |
| conv-49-q49 | MH | ENTITY_LINK | Evan's son and Evan | Soccer/health |
| conv-42-q4 | OD | TRUE_WRITE_MISS | Hairless cats or pigs | Decision filler |
| conv-48-q124 | SH | TRUE_WRITE_MISS | gathering information, videos | Sustainability projects |

## Takeaway

Do not treat WRITE_MISS=120 as "the regex compiler missed 120 facts." At least
one row is an oracle punctuation miss, and temporal/entity/paraphrase rows are
not compile-absent. Re-run the ledger with P3 ordering before allocating the
next compiler increment.
