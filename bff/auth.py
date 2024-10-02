from flask import Flask, request, redirect, url_for, make_response, jsonify, render_template
import requests
import os
import datetime

API_URL = os.getenv("API_URL", "http://localhost:8080")  

app = Flask(__name__)

def signup():
    if request.method == "GET":
        return render_template('sign_up.html')
    if request.method == "POST":
        try:
            # Convert form data to a dictionary
            data = {
                'email': request.form.get('email'),
                'password': request.form.get('password')
            }

            # Log the data to verify it before sending
            print(f"Sending data to API: {data}")

            # Send the request with JSON data
            r = requests.post(
                f'{API_URL}/signup',
                json=data,  
                headers={'Content-Type': 'application/json'}
            )

            response_json = r.json()

            if r.status_code == 201:
                return redirect(url_for("login_route"))
            elif r.status_code in (400, 409):
                error_message = response_json.get("message")
                return render_template('sign_up.html', error=error_message)
            elif r.status_code == 500:
                return make_response(jsonify({"error": response_json.get("message")}), 500)
            else:
                return make_response(jsonify({"error": "An unexpected error occurred"}), r.status_code)
        except requests.RequestException as e:
            return make_response(jsonify({"error": "Failed to connect to the external API", "details": str(e)}), 500)

def login():
    if request.method == "GET":
        return render_template("login.html")
    
    if request.method == "POST":
        # Convert form data to a dictionary
        data = {
            'email': request.form.get('email'),
            'password': request.form.get('password')
        }

        # Log the data to verify it before sending
        print(f"Sending data to API: {data}")

        # Send the request with JSON data
        try:
            r = requests.post(
                f'{API_URL}/login',
                json=data,
                headers={'Content-Type': 'application/json'}
            )
            
            if r.status_code == 200:
                # Extract token from API response
                userToken = r.json().get("token")
                
                # Set the token in a cookie with expiration date
                expire_date = datetime.datetime.now() + datetime.timedelta(days=1)
                resp = make_response(redirect("/"))  # Redirect to homepage or other URL
                resp.set_cookie("token", value=userToken, expires=expire_date)

                return resp
            
            elif r.status_code in [400, 404]:
                # Handle client errors (e.g., invalid login or user not found)
                error_message = r.json().get("message")
                return render_template('login.html', error=error_message)

            elif r.status_code == 500:
                # Handle server errors
                return make_response(jsonify({"error": "Internal server error"}), 500)

        except requests.RequestException as e:
            # Handle connection errors (e.g., API not reachable)
            return make_response(jsonify({"error": "Failed to connect to the API", "details": str(e)}), 500)

@app.route('/logout')
def logout():
    resp = make_response(redirect("/login"))
    resp.set_cookie('token', '', expires=0)
    return resp
