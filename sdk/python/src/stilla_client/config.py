from pydantic import BaseModel, validator, UUID4
import requests
from typing import Mapping, Any
import validators


class Config(BaseModel):
    """Stilla client

    Stilla client for Stilla API. Stores the authentication
    details for the client. Use the client to retrieve
    configuration.

    Attributes:
        base_url: The API URL
        api_key: The API key for the client.
        host_id: The host ID of the client.
    """

    base_url: str = "https://stilla.com"
    api_key: UUID4 = None
    host_id: UUID4 = None

    def __init__(self, **data):
        super().__init__(**data)

    def get_config(self, key: str) -> Mapping:
        """Retrieves configuration from Stilla API

        Args:
            key: The configuration key to retrieve.

        Returns:
            The mapping of the configuration.

        Raises:
            requests.exceptions.HTTPError: If the connection is not found.
            requests.Timeout: Connection timeout.
            requests.ConnectionError: A connection error occurred.
        """
        if key == "":
            raise ValueError("The key must not be empty.")
        try:
            headers = {
                "Authorization": f"Bearer {self.api_key}",
                "HostID": f"{self.host_id}",
            }
            r = requests.get(
                f"{self.base_url}/api/v1/host/{self.host_id}/config/{key}",
                headers=headers,
            )
            return {"data": r.json()}
        except requests.exceptions.HTTPError:
            return {"error": "configuration not found"}
        except requests.Timeout:
            return {"error": "connection timeout"}
        except requests.ConnectionError:
            return {"error": "connection error"}
        except Exception as e:
            return {"error": e}
        
    def get_config_value(self, key: str, config_key: str) -> Any:
        """_summary_

        Args:
            key (str): The configuration key to retrieve.

        Raises:
            ValueError: Value was not found in map.

        Returns:
            any: Returns the value in its format.
        """
        configuration = self.get_config(key)
        try:
            value = configuration[config_key]
            return value
        except ValueError:
            return {"error": "invalid key. value not found."}
        except Exception as e:
            return {"error": e}

    @validator("base_url")
    def valid_url(cls, v):
        validation = validators.url(v)
        if not validation:
            raise ValueError("must be a valid url")
        return v
