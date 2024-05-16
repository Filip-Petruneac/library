from flask import Flask, render_template, jsonify, request, url_for, send_from_directory, redirect
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
        data = request.get_json()

        response = requests.put(f"{API_URL}/authors/{author_id}", json=data)

        if response.status_code == 200:
            return jsonify(success=True)
        elif response.status_code == 400:
            return jsonify(success=False, error="Invalid data provided"), 400
        elif response.status_code == 404:
            return jsonify(success=False, error="Author not found"), 404
        else:
            return jsonify(success=False, error="Failed to update author"), 500

    except Exception as err:
        return jsonify(success=False, error=str(err)), 500
    

@app.route('/add_author', methods=['GET', 'POST'])
def add_author():
    if request.method == 'POST':
        firstname = request.form.get('firstname')
        lastname = request.form.get('lastname')
        photo = request.form.get('photo')

        data = {
            'firstname': firstname,
            'lastname': lastname,
            'photo': photo
        }

        try:
            response = requests.post(f"{API_URL}/authors/new", json=data)
            if response.status_code == 201:
                return redirect(url_for('/authors'))
            else:
                if response.content:
                    error_message = response.json().get('error', 'Failed to add author')
                else:
                    error_message = 'Empty response from the API'
                app.logger.error(f"Failed to add author: {error_message}")
                return jsonify(success=False, error=error_message), 500
            
        except Exception as err:
            app.logger.error(f"Failed to add author: {err}")
            return jsonify(success=False, error=str(err)), 500
        
    return render_template('add_form.html')



@app.route('/css/<path:filename>')
def serve_css(filename):
    return send_from_directory('static/css', filename)

@app.route('/js/<path:filename>')
def serve_js(filename):
    return send_from_directory('static/js', filename)

if __name__ == '__main__':
    app.run(debug=True, port=5000)
