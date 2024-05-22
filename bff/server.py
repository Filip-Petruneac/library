from flask import Flask, render_template, jsonify, request, url_for, send_from_directory, redirect
import requests
from flask import jsonify
import os
from werkzeug.utils import secure_filename

UPLOAD_FOLDER = './photos'
ALLOWED_EXTENSIONS = set(['txt', 'pdf', 'png', 'jpg', 'jpeg', 'gif'])

app = Flask(__name__)

app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER

API_URL = "http://localhost:8080"  

@app.route('/')
def index():
    # Apelează funcția get_books() pentru a obține cărțile de la API
    try:
        response = requests.get(f"{API_URL}/books")
        if response.status_code != 200:
            return "Error fetching books from API", 400
        
        books = response.json()
        return render_template('books.html', books=books)  # Trimite datele către șablonul HTML
    
    except Exception as err:
        return str(err), 500
    
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


@app.route('/books', methods=['GET'])
def get_books():
    try:
        response = requests.get(f"{API_URL}/books")
        if response.status_code != 200:
            return "Error fetching authors from API", 400
        
        books = response.json()
        return jsonify(books)  
    
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


@app.route('/book-details/<int:book_id>')
def book_details(book_id):
    try:
        response = requests.get(f"{API_URL}/books/{book_id}")
        if response.status_code != 200:
            return "Error fetching book details from API", 400
        
        book = response.json()
        return render_template('book_details.html', book=book)
    
    except Exception as err:
        return str(err), 500

@app.route('/update_author_form.html', methods=['GET'])
def update_author_form():
    return render_template('update_author_form.html')

@app.route('/author/<int:author_id>', methods=['PUT'])
def update_author(author_id):
    try:
        data = request.get_json()
        response = requests.put(f"{API_URL}/authors/{author_id}", json=data)
        if response.status_code != 200:
            return jsonify(success=False, error="Error updating author"), 400
        return jsonify(success=True)

    except Exception as err:
        return jsonify(success=False, error=str(err)), 500
    
@app.route('/book/<int:book_id>', methods=['PUT'])
def update_book(book_id):
    try:
        data = request.get_json()
        response = requests.put(f"{API_URL}/books/{book_id}", json=data)
        if response.status_code != 200:
            return jsonify(success=False, error="Error updating book"), 400
        return jsonify(success=True)

    except Exception as err:
        return jsonify(success=False, error=str(err)), 500
    
@app.route('/add_author', methods=['GET', 'POST'])
def add_author():
    if request.method == 'POST':
        firstname = request.form.get('firstname')
        lastname = request.form.get('lastname')
        photo = request.files['photo']

        if photo:
            filename = secure_filename(photo.filename)
            photo_path = os.path.join(app.config['UPLOAD_FOLDER'],filename)
            photo.save(photo_path)
            photo_path = f'photos/{filename}'

        else:
            photo.filename = None
            photo_path = None

        data = {
            'firstname': firstname,
            'lastname': lastname,
            'photo': photo_path
        }

        try:
            response = requests.post(f"{API_URL}/authors/new", json=data)
            if response.status_code == 201:
                if photo_path and os.path.exists(os.path.join(app.config['UPLOAD_FOLDER'], filename)):
                    os.remove(os.path.join(app.config['UPLOAD_FOLDER'], filename))
                return redirect(url_for("get_authors"))
            else:
                if response.content:
                    print(response)
                    error_message = response.json().get('error', 'Failed to add author')
                else:
                    error_message = 'Empty response from the API'
                    print("Empty response:", response.content)

                app.logger.error(f"Failed to add author: {error_message}")
                return jsonify(success=False, error=error_message), 500
            
        except Exception as err:
            app.logger.error(f"Failed to add author: {err}")
            return jsonify(success=False, error=str(err)), 500
        
    return render_template('add_author_form.html')

@app.route('/css/<path:filename>')
def serve_css(filename):
    return send_from_directory('static/css', filename)

@app.route('/js/<path:filename>')
def serve_js(filename):
    return send_from_directory('static/js', filename)

if __name__ == '__main__':
    app.run(debug=True, port=5000)
