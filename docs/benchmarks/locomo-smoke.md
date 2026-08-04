# LOCOMO smoke — P0 de-overfit re-baseline

**Commit:** `ece8d52` · gpt-oss · top_k=30 · async · **no LOCOMO surface-forms in product**

| Line | Overall | multi-hop | temporal | open |
| --- | ---: | ---: | ---: | ---: |
| Pre-W1 peak (with hacks) | 19/30 | 5/10 | 11–13/16 | 3–4/4 |
| **P0 honest baseline** | **16/30** | **4/10** | 9/16 | 3/4 |
| Mem0 same-pin (prior) | 12/30 | 6/10 | 2/16 | 4/4 |

OpMem: **12/12** (non-regress).

This drop vs 19/30 is the expected honest starting line after W1 (master-plan §1.2 / §9.1).
Wins back via W2–W4 typed atoms + enumeration + `/recall`, not by restoring regex answer keys.
