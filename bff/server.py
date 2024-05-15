from flask import Flask, render_template, jsonify, request, url_for, send_from_directory
import requests

app = Flask(__name__)

API_URL = "http://localhost:8080"  

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

@app.route('/author/<int:author_id>', methods=['DELETE'])
def delete_author(author_id):
    try:
        response = requests.delete(f"{API_URL}/authors/{author_id}")
        if response.status_code != 200:
            return jsonify(success=False), 400
        
        return jsonify(success=True)
    
    except Exception as err:
        return jsonify(success=False, error=str(err)), 500

@app.route('/author/<int:author_id>', methods=['PUT'])
def update_author(author_id):
    try:
        # ia/extrage t
        data = request.get_json()

        # apeleaza api-ul sa faca update la autor
        response = requests.put(f"{API_URL}/authors/{author_id}", json=data)

        # verifica raspunsul din ai
        if response.status_code != 200:
            return jsonify(success=False, error="Error updating author"), 400

        return jsonify(success=True)

    except Exception as err:
        return jsonify(success=False, error=str(err)), 500

@app.route('/css/<path:filename>')
def serve_css(filename):
    return send_from_directory('static/css', filename)

@app.route('/js/<path:filename>')
def serve_js(filename):
    return send_from_directory('static/js', filename)

if __name__ == '__main__':
    app.run(debug=True, port=5000)
