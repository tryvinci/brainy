# OpMem non-reg — Gate 0 staging — 2026-08-11

**Target:** Render `brainy-api-staging` @ `9bad8987483a`  
**Result:** **13/13 passed**

```
{
  "benchmark": "opmem-v0",
  "systems": [
    "brainy"
  ],
  "summary": {
    "brainy": {
      "overall": "13/13",
      "by_category": {
        "correction": "3/3",
        "idempotency": "1/1",
        "isolation": "3/3",
        "staleness": "3/3",
        "suppression": "3/3"
      }
    }
  },
  "infrastructure_errors": 0,
  "results": [
    {
      "task": "cor01_basic_revision",
      "category": "correction",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "cor02_correction_stickiness",
      "category": "correction",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "cor03_revised_retrievable",
      "category": "correction",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "dup01_idempotent_remember",
      "category": "idempotency",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "iso01_subject_isolation",
      "category": "isolation",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "iso02_tenant_isolation",
      "category": "isolation",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "iso03_forget_isolated",
      "category": "isolation",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "sup01_basic_forget",
      "category": "suppression",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "sup02_targeted_forget",
      "category": "suppression",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "sup03_durable_forget",
      "category": "suppression",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "upd01_stale_fact",
      "category": "staleness",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "upd02_preference_change",
      "category": "staleness",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    },
    {
      "task": "upd03_state_supersession",
      "category": "staleness",
      "systems": {
        "brainy": {
          "passed": true,
          "failures": []
        }
      }
    }
  ]
}

```
