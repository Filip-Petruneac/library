from flask import Flask, render_template, jsonify, request
import requests

app = Flask(__name__)

API_URL = "http://localhost:8080"  # Adresa la care rulează serverul API

@app.route('/authors', methods=['GET'])
def get_authors():
    try:
        response = requests.get(f"{API_URL}/authors")
        if response.status_code != 200:
            return "Error fetching authors from API", 500
        
        authors = response.json()
        return render_template('authors.html', authors=authors)
    
    except Exception as e:
        return str(e), 500

@app.route('/delete-author/<int:author_id>', methods=['DELETE'])
def delete_author(author_id):
    try:
        response = requests.delete(f"{API_URL}/delete-author", json={"id": author_id})
        if response.status_code != 200:
            return jsonify(success=False), 500
        
        return jsonify(success=True)
    
    except Exception as e:
        return jsonify(success=False, error=str(e)), 500

if __name__ == '__main__':
    app.run(debug=True, port=5000)
