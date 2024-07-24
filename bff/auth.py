from functools import wraps
from flask import request, redirect, url_for, make_response, jsonify, render_template
import requests
import datetime

API_URL = "http://localhost:8081"

def is_authenticated():
    user_id = request.cookies.get('authenticatedUserID')
    return user_id is not None

def login_required(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        if not is_authenticated():
            return redirect(url_for('login_route'))
        return f(*args, **kwargs)
    return decorated_function

def signup():
    if request.method == "GET":
        return render_template('sign_up.html')
    if request.method == "POST":
        try:
            r = requests.post(f'{API_URL}/singup', request.form, headers=request.headers)
            response_json = r.json()

            if r.status_code == 201:
                return redirect(url_for("login_route"))
            elif r.status_code == 400:
                error_message = response_json.get("message")
                return render_template('sign_up.html', error=error_message)
            elif r.status_code == 409:
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
        try:
            r = requests.post(f'{API_URL}/login', request.form, headers=request.headers)
            response_json = r.json()

            if r.status_code == 200:
                id = response_json.get("existingUserID")
                expire_date = datetime.datetime.now()
                expire_date = expire_date + datetime.timedelta(days=1)

                resp = make_response(redirect("/"))  
                resp.set_cookie("authenticatedUserID", value=str(id), expires=expire_date) 

                return resp
            elif r.status_code == 400:
                error_message = response_json.get("message")
                return render_template('login.html', error=error_message)
            elif r.status_code == 404:
                error_message = response_json.get("message")
                return render_template('login.html', error=error_message)
            elif r.status_code == 500:
                return make_response(jsonify({"error": response_json.get("message")}), 500)
            else:
                return make_response(jsonify({"error": "An unexpected error occurred"}), r.status_code)
        except requests.RequestException as e:
            return make_response(jsonify({"error": "Failed to connect to the external API", "details": str(e)}), 500)

def logout():
    resp = make_response(redirect("/"))
    resp.set_cookie('authenticatedUserID', '', expires=0)
    return resp