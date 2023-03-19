"""Stilla Client SDK used to retrieve configuration from the Stilla API.

The Stilla API allows clients to store and retrieve configuration. This
Client SDK allows a client to retrieve configuration for use in a service
or script.

Typical usage example:

  client = Config(api_key="d6845fe3-7054-4323-8ce3-468f25f5b52d", host_id="c0bddbd3-5ccb-4376-901c-6e3cf87152fa")
  config = client.get_config(key="test")
"""
