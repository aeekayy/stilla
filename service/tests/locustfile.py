from locust import HttpUser, task, between
from locust.clients import HttpSession
import os

class ConfigUser(HttpUser):
    wait_time = between(0.5, 3)

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.api_client = HttpSession(
            base_url=os.getenv("API_HOST", "http://localhost:8080"), request_event=self.client.request_event, user=self
        )

    @task
    def get_config(self):
        token = os.getenv("API_TOKEN", "cfacd739-4a13-47ae-82c3-13d6d7ffeb2e")
        hostID = os.getenv("HOST_ID", "9923d21c-dbac-421d-a31a-649a849d4c85")
        
        bearer_token = f"Bearer {token}"
        self.client.get(
            url=f"/api/v1/host/{hostID}/config/kubernetes",
            headers={ "Authorization": bearer_token, "HostID": hostID }
        )