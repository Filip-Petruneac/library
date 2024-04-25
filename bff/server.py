from flask import Flask, render_template, jsonify, request
import requests

app = Flask(__name__)

API_URL = "http://localhost:8080"  # Adresa la care rulează serverul API

@app.route('/authors', methods=['GET'])
def get_authors():
    try:
        response = requests.get(f"{API_URL}/authors")
        if response.status_code != 200:
            return "Error fetching authors from API", 400
        
        authors = response.json()
        return render_template('authors.html', authors=authors)
    
    except Exception as err:
        return str(err), 500

@app.route('/delete-author/<int:author_id>', methods=['DELETE'])
def delete_author(author_id):
    try:
        response = requests.delete(f"{API_URL}/authors/" + author_id)
        if response.status_code != 200:
            return jsonify(success=False), 400
        
        return jsonify(success=True)
    
    except Exception as err:
        return jsonify(success=False, error=str(err)), 500

if __name__ == '__main__':
    app.run(debug=True, port=5000)
