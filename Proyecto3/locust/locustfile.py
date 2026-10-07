from datetime import datetime, timezone
import random
from locust import HttpUser, between, task


class MilitaryReportUser(HttpUser):
    wait_time = between(0.1, 0.5)

    @task
    def send_report(self):
        payload = {
            "country": random.choice(["USA", "RUS", "CHN", "ESP", "GTM"]),
            "warplanes_in_air": random.randint(0, 50),
            "warships_in_water": random.randint(0, 30),
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }
        self.client.post("/grpc-2023204", json=payload, name="/grpc-2023204")
