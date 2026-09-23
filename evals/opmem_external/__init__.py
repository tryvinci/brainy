"""OpMem external memory lane adapters (Phase 1: dry-run + instrumentation)."""

from opmem_external.langmem_store import LangMemLaneAdapter
from opmem_external.letta_appserver import LettaAppServerLaneAdapter
from opmem_external.mem0_oss import Mem0OSSLaneAdapter
from opmem_external.zep_cloud import ZepCloudLaneAdapter

LANE_ADAPTERS = {
    "mem0-oss": Mem0OSSLaneAdapter,
    "zep": ZepCloudLaneAdapter,
    "letta": LettaAppServerLaneAdapter,
    "langmem": LangMemLaneAdapter,
}

__all__ = [
    "LANE_ADAPTERS",
    "LangMemLaneAdapter",
    "LettaAppServerLaneAdapter",
    "Mem0OSSLaneAdapter",
    "ZepCloudLaneAdapter",
]
