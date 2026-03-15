"""
layer2/lake.py
Meteorium Engine — LAYER 2: Raw Telemetry Lake
Immutable object store for all raw satellite/agency files.
Manifest index tracks every deposited file with full provenance.
Files are NEVER modified after deposit — preprocessing works from copies.
"""

import hashlib
import json
import logging
import sqlite3
import time
from contextlib import contextmanager
from datetime import datetime, timezone, timedelta
from pathlib import Path
from typing import Optional

from core.config import LAKE_DIR, STORE_DIR
from core.models import LakeEntry, IngestRecord

logger = logging.getLogger("meteorium.layer2")


class DataLake:
    """
    Layer 2: Immutable raw telemetry storage with manifest index.

    Structure:
        lake/
            firms/          ← NASA FIRMS CSV files
            era5/           ← ERA5 NetCDF + OpenMeteo JSON
            gsmap/          ← JAXA GSMaP precipitation
            carbon/         ← Carbon Monitor JSON
            cmip6/          ← CMIP6 scenario NetCDF
            srtm/           ← SRTM elevation GeoTIFF

        store/
            manifest.db     ← SQLite manifest of all lake files
    """

    MANIFEST_PATH = STORE_DIR / "manifest.db"

    def __init__(self):
        self._init_manifest()
        logger.info(f"[Lake] Initialized. Root: {LAKE_DIR}")

    def _init_manifest(self):
        """Create manifest database if not exists."""
        with self._conn() as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS lake_manifest (
                    lake_id      TEXT PRIMARY KEY,
                    source_id    TEXT NOT NULL,
                    variable     TEXT,
                    file_path    TEXT NOT NULL,
                    bbox_json    TEXT,
                    time_start   TEXT,
                    time_end     TEXT,
                    resolution   REAL,
                    file_hash    TEXT,
                    file_size_mb REAL,
                    deposited_at TEXT NOT NULL,
                    expires_at   TEXT
                )
            """)
            conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_source_deposited
                ON lake_manifest(source_id, deposited_at DESC)
            """)
            conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_time_range
                ON lake_manifest(time_start, time_end)
            """)
            conn.commit()

    @contextmanager
    def _conn(self):
        conn = sqlite3.connect(self.MANIFEST_PATH)
        conn.row_factory = sqlite3.Row
        try:
            yield conn
        finally:
            conn.close()

    # ── Deposit ───────────────────────────────────────────────────────────────

    def deposit(self, record: IngestRecord, variable: str = "") -> str:
        """
        Register a completed ingestion record in the manifest.
        The physical file is already written by the worker.
        Returns the lake_id.
        """
        lake_id = self._generate_id(record.source_id, record.ingested_at)

        with self._conn() as conn:
            conn.execute("""
                INSERT OR REPLACE INTO lake_manifest
                (lake_id, source_id, variable, file_path, bbox_json,
                 time_start, time_end, resolution, file_hash,
                 file_size_mb, deposited_at, expires_at)
                VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
            """, (
                lake_id,
                record.source_id,
                variable,
                record.file_path,
                json.dumps(list(record.bbox)),
                record.time_start.isoformat(),
                record.time_end.isoformat(),
                1.0,
                record.file_hash,
                record.file_size_mb,
                record.ingested_at.isoformat(),
                None,
            ))
            conn.commit()

        logger.debug(f"[Lake] Deposited {lake_id} ({record.source_id})")
        return lake_id

    # ── Query ─────────────────────────────────────────────────────────────────

    def latest(
        self,
        source_id: str,
        max_age_hours: float = 24.0,
    ) -> Optional[LakeEntry]:
        """Get most recent file for a source within max_age_hours."""
        since = (datetime.now(timezone.utc) - timedelta(hours=max_age_hours)).isoformat()

        with self._conn() as conn:
            row = conn.execute("""
                SELECT * FROM lake_manifest
                WHERE source_id = ? AND deposited_at >= ?
                ORDER BY deposited_at DESC LIMIT 1
            """, (source_id, since)).fetchone()

        if not row:
            return None

        return self._row_to_entry(row)

    def list_files(
        self,
        source_id: Optional[str] = None,
        since_hours: Optional[float] = None,
    ) -> list[LakeEntry]:
        """List files in the manifest."""
        query = "SELECT * FROM lake_manifest WHERE 1=1"
        params = []

        if source_id:
            query += " AND source_id = ?"
            params.append(source_id)

        if since_hours:
            since = (datetime.now(timezone.utc) - timedelta(hours=since_hours)).isoformat()
            query += " AND deposited_at >= ?"
            params.append(since)

        query += " ORDER BY deposited_at DESC"

        with self._conn() as conn:
            rows = conn.execute(query, params).fetchall()

        return [self._row_to_entry(r) for r in rows]

    def stats(self) -> dict:
        """Lake storage statistics."""
        with self._conn() as conn:
            total = conn.execute(
                "SELECT COUNT(*) as n, SUM(file_size_mb) as mb FROM lake_manifest"
            ).fetchone()
            by_source = conn.execute("""
                SELECT source_id, COUNT(*) as n,
                       MAX(deposited_at) as latest
                FROM lake_manifest
                GROUP BY source_id
                ORDER BY latest DESC
            """).fetchall()

        return {
            "total_files":   total["n"] or 0,
            "total_size_mb": round(total["mb"] or 0, 2),
            "by_source":     [
                {
                    "source_id": r["source_id"],
                    "file_count": r["n"],
                    "latest":    r["latest"],
                }
                for r in by_source
            ],
        }

    def purge_expired(self) -> int:
        """Remove expired entries from manifest (files are deleted too)."""
        now = datetime.now(timezone.utc).isoformat()
        with self._conn() as conn:
            rows = conn.execute(
                "SELECT file_path FROM lake_manifest WHERE expires_at < ?", (now,)
            ).fetchall()
            count = len(rows)
            for row in rows:
                try:
                    Path(row["file_path"]).unlink(missing_ok=True)
                except Exception:
                    pass
            conn.execute(
                "DELETE FROM lake_manifest WHERE expires_at < ?", (now,)
            )
            conn.commit()
        if count:
            logger.info(f"[Lake] Purged {count} expired entries")
        return count

    # ── Helpers ───────────────────────────────────────────────────────────────

    @staticmethod
    def _generate_id(source_id: str, ts: datetime) -> str:
        ts_str = ts.strftime("%Y%m%d%H%M%S%f")
        return f"{source_id}_{ts_str}"

    @staticmethod
    def _row_to_entry(row: sqlite3.Row) -> LakeEntry:
        def _dt(s):
            if not s:
                return None
            try:
                return datetime.fromisoformat(s)
            except ValueError:
                return datetime.now(timezone.utc)

        return LakeEntry(
            lake_id      = row["lake_id"],
            source_id    = row["source_id"],
            variable     = row["variable"] or "",
            file_path    = row["file_path"],
            bbox         = row["bbox_json"] or "[]",
            time_start   = _dt(row["time_start"]) or datetime.now(timezone.utc),
            time_end     = _dt(row["time_end"])   or datetime.now(timezone.utc),
            resolution   = row["resolution"] or 1.0,
            file_hash    = row["file_hash"] or "",
            file_size_mb = row["file_size_mb"] or 0.0,
            deposited_at = _dt(row["deposited_at"]) or datetime.now(timezone.utc),
            expires_at   = _dt(row["expires_at"]),
        )
