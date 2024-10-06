#!/usr/bin/env python3

import requests
import uuid
import random
import jwt  # Import the jwt library
from datetime import datetime, timedelta

# Define the base URL
host = "localhost:10000"
base_url = f"http://{host}/v1"

# Function to handle user login
def login(email, password):
    url = f"{base_url}/users/login"
    headers = {"Content-Type": "application/json"}
    data = {"email": email, "password": password}

    # Send POST request for user login
    response = requests.post(url, json=data, headers=headers)

    # Check if login was successful
    if response.status_code == 200:
        print("Login successful")
        return response.json()  # Return the response for further use
    else:
        print(f"Login failed: {response.status_code}, {response.text}")
        return None

# Function to read UUIDs from a file
def read_uuids_from_file(file_path):
    with open(file_path, 'r') as file:
        return [line.strip() for line in file if line.strip()]

# Read track UUIDs from a file
uuids_track = read_uuids_from_file('tracks.txt')

# Read playlist UUIDs from a file
uuids_playlist = read_uuids_from_file('playlist.txt')

# Authenticate the user
user_data = login("a@a.com", "1")

# Check if authentication was successful
if user_data:
    # Retrieve the token from the login response
    token = user_data.get("jwt_token")
    print(f"user_data: {user_data}")

    # Prepare the headers based on the required format
    headers = {
        "Authorization": f"Bearer {token}",
        "User-Agent": "IntelliJ HTTP Client/GoLand 2024.1.3",
        "Accept-Encoding": "br, deflate, gzip, x-gzip",
        "Accept": "*/*",
        "Cookie": f"jwt={token}; refresh_token=<your_refresh_token_here>; session=<your_session_here>",
        "Content-Length": "0"  # This is for the GET request where no body content is sent
    }


    # Iterate through the arrays and send POST requests to add tracks to playlists
    for playlist_uuid in uuids_playlist:
        for track_uuid in uuids_track:
            url = f"{base_url}/playlist/{playlist_uuid}/{track_uuid}"
            response = requests.post(url, headers=headers)  # Include the token in the headers

            # Check for a successful response
            if response.status_code == 200:
                print(f"Successfully added track {track_uuid} to playlist {playlist_uuid}: {response.json()}")
            else:
                print(f"Failed to add track {track_uuid} to playlist {playlist_uuid}: {response.status_code}, {response.text}")


    # Randomly select a track from uuids_track for playlist_uuid[0]
    random_track_uuid = random.choice(uuids_track)
    playlist_uuid_0 = uuids_playlist[0]
    url = f"{base_url}/playlist/{playlist_uuid_0}/{random_track_uuid}"

    # Send the request to add a random track to the first playlist
    response = requests.post(url, headers=headers)

    # Check for a successful response
    if response.status_code == 200:
        print(f"Successfully added random track {random_track_uuid} to playlist {playlist_uuid_0}: {response.json()}")
    else:
        print(f"Failed to add random track {random_track_uuid} to playlist {playlist_uuid_0}: {response.status_code}, {response.text}")

    # Add playlist to playlist
    playlist_uuid_0 = uuids_playlist[0]
    playlist_uuid_to = uuids_playlist[1]
    url = f"{base_url}/playlist/{playlist_uuid_0}/{playlist_uuid_to}"

    # Send the request to add a random track to the first playlist
    response = requests.post(url, headers=headers)

    # Check for a successful response
    if response.status_code == 200:
        print(f"Successfully added random track {playlist_uuid_to} to playlist {playlist_uuid_0}: {response.json()}")
    else:
        print(f"Failed to add random track {playlist_uuid_to} to playlist {playlist_uuid_0}: {response.status_code}, {response.text}")
