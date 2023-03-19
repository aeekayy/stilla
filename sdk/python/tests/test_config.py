from stilla_client.config import Config
import pytest
import requests
import requests_mock


class TestConfig:
    client: Config = None

    def __init__(self):
        self.client = Config(
            api_key="a49da419-3899-46a0-b43a-5bf4a26d3cca",
            host_id="c2f0d0a2-af9b-4b40-96d8-86865eb958c2",
        )


def test_cls_no_base_url():
    with pytest.raises(ValueError) as e_info:
        client = Config(base_url="")


def test_cls_invalid_url():
    with pytest.raises(ValueError) as e_info:
        client = Config(base_url="foobar")


@requests_mock.Mocker(kw="mock")
def test_get_config_pass_multi(**kwargs):
    tc = TestConfig()
    keys = ["gke", "eks"]
    resp = [
        {"gke:name": "test", "gke:region": "us-central1"},
        {"eks:name": "test", "eks:region": "us-east-1"},
    ]
    expected = [
        {"data": {"gke:name": "test", "gke:region": "us-central1"}},
        {"data": {"eks:name": "test", "eks:region": "us-east-1"}},
    ]
    for i in range(0, 2):
        kwargs["mock"].register_uri(
            "GET",
            f"{tc.client.base_url}/api/v1/host/{tc.client.host_id}/config/{keys[i]}",
            request_headers={
                "Authorization": f"Bearer {tc.client.api_key}",
                "HostID": f"{tc.client.host_id}",
            },
            json=resp[i],
        )
        config = tc.client.get_config(key=f"{keys[i]}")
        assert expected[i] == config


@requests_mock.Mocker(kw="mock")
def test_get_config_pass_single(**kwargs):
    tc = TestConfig()
    resp = {"enabled": "true"}
    expected = {"data": {"enabled": "true"}}
    kwargs["mock"].register_uri(
        "GET",
        f"{tc.client.base_url}/api/v1/host/{tc.client.host_id}/config/cron",
        request_headers={
            "Authorization": f"Bearer {tc.client.api_key}",
            "HostID": f"{tc.client.host_id}",
        },
        json=resp,
    )
    config = tc.client.get_config(key="cron")
    assert expected == config


@requests_mock.Mocker(kw="mock")
def test_get_config_fail_empty_key(**kwargs):
    tc = TestConfig()
    resp = {"gke:name": "test", "gke:region": "us-central1"}
    kwargs["mock"].register_uri(
        "GET",
        f"{tc.client.base_url}/api/v1/host/{tc.client.host_id}/config/gke",
        request_headers={
            "Authorization": f"Bearer {tc.client.api_key}",
            "HostID": f"{tc.client.host_id}",
        },
        json=resp,
    )
    with pytest.raises(ValueError) as e_info:
        config = tc.client.get_config(key="")


# TODO Add exception tests
