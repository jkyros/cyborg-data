#!/usr/bin/env python3
"""
Verify stable_id fields are present in the production GCS index.

Loads the comprehensive index from GCS and checks that entities
have stable_id values populated by the upstream identity tracker.

Requires: pip install google-cloud-storage
Auth: gcloud auth application-default login
"""

from datetime import timedelta

from orgdatacore import GCSConfig, GCSDataSource, Service


def main():
    config = GCSConfig(
        bucket="resolved-org",
        object_path="orgdata/comprehensive_index_dump.json",
        project_id="openshift-crt",
        check_interval=timedelta(minutes=5),
    )

    print(f"Loading from gs://{config.bucket}/{config.object_path} ...")
    source = GCSDataSource(config)
    service = Service()

    try:
        service.load_from_data_source(source)
    except Exception as e:
        print(f"Failed to load: {e}")
        print("  gcloud auth application-default login")
        return

    version = service.get_version()
    print(f"Loaded: {version.employee_count} employees, {version.org_count} orgs")
    print()

    entity_types = [
        ("Teams", service.get_all_teams),
        ("Orgs", service.get_all_orgs),
        ("Pillars", service.get_all_pillars),
        ("Team Groups", service.get_all_team_groups),
    ]

    total_entities = 0
    total_with_id = 0
    total_without_id = 0

    for label, get_all in entity_types:
        entities = get_all()
        with_id = [e for e in entities if e.stable_id]
        without_id = [e for e in entities if not e.stable_id]

        total_entities += len(entities)
        total_with_id += len(with_id)
        total_without_id += len(without_id)

        print(f"{label}: {len(entities)} total, {len(with_id)} with stable_id, {len(without_id)} without")

        if with_id:
            sample = with_id[:3]
            for e in sample:
                print(f"  {e.stable_id}  {e.name}")

        if without_id and len(without_id) <= 5:
            for e in without_id:
                print(f"  (no id)   {e.name}")
        elif without_id:
            print(f"  ... {len(without_id)} entities missing stable_id")

        print()

    print("=" * 50)
    print(f"Total: {total_entities} entities, {total_with_id} with stable_id, {total_without_id} without")
    print()

    # Test reverse lookup
    if total_with_id > 0:
        teams = service.get_all_teams()
        sample = next((t for t in teams if t.stable_id), None)
        if sample:
            print("Reverse lookup test:")
            print(f"  Looking up stable_id={sample.stable_id} ...")

            result = service.get_entity_by_stable_id(sample.stable_id)
            if result:
                print(f"  get_entity_by_stable_id -> name={result.name}, type={result.type}")
            else:
                print("  get_entity_by_stable_id -> None (UNEXPECTED)")

            team = service.get_team_by_stable_id(sample.stable_id)
            if team:
                print(f"  get_team_by_stable_id   -> {team.name} (matches: {team.name == sample.name})")
            else:
                print("  get_team_by_stable_id   -> None (UNEXPECTED)")

    if total_with_id == 0:
        print("WARNING: No entities have stable_id set. The identity map may not be")
        print("wired into the production indexer yet (PR #292 adds the --identity-map flag).")


if __name__ == "__main__":
    main()
