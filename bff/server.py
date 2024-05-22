from flask import Flask, render_template, jsonify, request, url_for, send_from_directory, redirect
import requests
from flask import jsonify
import os
from werkzeug.utils import secure_filename


app = Flask(__name__)

app.config['UPLOAD_FOLDER'] = '/home/filip/Downloads'

API_URL = "http://localhost:8080"  

@app.route('/')
def index():
    try:
        response = requests.get(f"{API_URL}/books")
        if response.status_code != 200:
            return "Error fetching books from API", 400
        
        books = response.json()
        return render_template('books.html', books=books)  
    
    except Exception as err:
        return str(err), 500
    
@app.route('/books')
def books():
    try:
        response = requests.get(f"{API_URL}/books")
        if response.status_code != 200:
            return "Error fetching books from API", 400

        books = response.json()
        return render_template('books.html', books=books)

    except Exception as err:
        return str(err), 500

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
    
@app.route('/css/<path:filename>')
def serve_css(filename):
    return send_from_directory('static/css', filename)

@app.route('/js/<path:filename>')
def serve_js(filename):
    return send_from_directory('static/js', filename)

if __name__ == '__main__':
    app.run(debug=True, port=5000)
