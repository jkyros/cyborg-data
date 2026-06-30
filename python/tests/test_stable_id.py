"""Tests for stable ID lookup functionality."""

import pytest

from orgdatacore import Service, StableIDResult


class TestGetEntityByStableID:
    """Tests for generic entity lookup by stable ID."""

    @pytest.mark.parametrize(
        "stable_id,expected_name,expected_type",
        [
            ("a1b2c3d4", "test-team", "team"),
            ("e5f6a7b8", "platform-team", "team"),
            ("c9d0e1f2", "test-org", "org"),
            ("a3b4c5d6", "platform-org", "org"),
            ("e7f8a9b0", "engineering", "pillar"),
            ("c1d2e3f4", "backend-teams", "team_group"),
            ("00000000", None, None),
            ("", None, None),
        ],
    )
    def test_get_entity_by_stable_id(
        self,
        service: Service,
        stable_id: str,
        expected_name: str | None,
        expected_type: str | None,
    ):
        result = service.get_entity_by_stable_id(stable_id)
        if expected_name is None:
            assert result is None
        else:
            assert result is not None
            assert result.name == expected_name
            assert result.type == expected_type


class TestGetTeamByStableID:
    """Tests for team lookup by stable ID."""

    @pytest.mark.parametrize(
        "stable_id,expected_name",
        [
            ("a1b2c3d4", "test-team"),
            ("e5f6a7b8", "platform-team"),
            ("c9d0e1f2", None),  # org stable ID
            ("00000000", None),
            ("", None),
        ],
    )
    def test_get_team_by_stable_id(
        self,
        service: Service,
        stable_id: str,
        expected_name: str | None,
    ):
        result = service.get_team_by_stable_id(stable_id)
        if expected_name is None:
            assert result is None
        else:
            assert result is not None
            assert result.name == expected_name


class TestGetOrgByStableID:
    """Tests for org lookup by stable ID."""

    @pytest.mark.parametrize(
        "stable_id,expected_name",
        [
            ("c9d0e1f2", "test-org"),
            ("a3b4c5d6", "platform-org"),
            ("a1b2c3d4", None),  # team stable ID
            ("00000000", None),
        ],
    )
    def test_get_org_by_stable_id(
        self,
        service: Service,
        stable_id: str,
        expected_name: str | None,
    ):
        result = service.get_org_by_stable_id(stable_id)
        if expected_name is None:
            assert result is None
        else:
            assert result is not None
            assert result.name == expected_name


class TestGetPillarByStableID:
    """Tests for pillar lookup by stable ID."""

    @pytest.mark.parametrize(
        "stable_id,expected_name",
        [
            ("e7f8a9b0", "engineering"),
            ("a1b2c3d4", None),  # team stable ID
            ("00000000", None),
        ],
    )
    def test_get_pillar_by_stable_id(
        self,
        service: Service,
        stable_id: str,
        expected_name: str | None,
    ):
        result = service.get_pillar_by_stable_id(stable_id)
        if expected_name is None:
            assert result is None
        else:
            assert result is not None
            assert result.name == expected_name


class TestGetTeamGroupByStableID:
    """Tests for team group lookup by stable ID."""

    @pytest.mark.parametrize(
        "stable_id,expected_name",
        [
            ("c1d2e3f4", "backend-teams"),
            ("a1b2c3d4", None),  # team stable ID
            ("00000000", None),
        ],
    )
    def test_get_team_group_by_stable_id(
        self,
        service: Service,
        stable_id: str,
        expected_name: str | None,
    ):
        result = service.get_team_group_by_stable_id(stable_id)
        if expected_name is None:
            assert result is None
        else:
            assert result is not None
            assert result.name == expected_name
