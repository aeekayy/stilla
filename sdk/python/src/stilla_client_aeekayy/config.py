from pydantic import BaseModel
import uuid
from typing import Mapping
import requests

class Config(BaseModel):
    base_url: str = "https://stilla.com"
    api_key: uuid.UUID = None
    host_id: uuid.UUID = None

    def get_config(self, key: str) -> Mapping:
        try:
            headers = {'Authorization': f'Bearer {self.api_key}', 'HostID': f'{self.host_id}'}
            r = requests.get(f'{self.base_url}/api/v1/host/{self.host_id}/config/{key}')

            if r.status_code != 200:
                return {}

            return r.json()