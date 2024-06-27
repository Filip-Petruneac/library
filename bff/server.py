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
        app.logger.debug(f"Book details: {book}")  # Log the book details
        return render_template('book_details.html', book=book)
    except Exception as err:
        return str(err), 500
    
@app.route('/subscribers', methods=['GET'])
def get_subscribers():
    try:
        response = requests.get(f"{API_URL}/subscribers")
        if response.status_code == 200:
            subscribers = response.json()
            return render_template('subscribers.html', subscribers=subscribers)
        else:
            error_message = response.json().get('error', 'Failed to retrieve subscribers')
            app.logger.error(f"Failed to retrieve subscribers: {error_message}")
            return jsonify(success=False, error=error_message), 500
    except Exception as err:
        app.logger.error(f"Failed to retrieve subscribers: {err}")
        return jsonify(success=False, error=str(err)), 500
    
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

@app.route('/book/<int:book_id>', methods=['DELETE'])
def delete_book(book_id):
    try:
        response = requests.delete(f"{API_URL}/books/{book_id}")
        if response.status_code != 200:
            return jsonify(success=False), 400
        return jsonify(success=True)
    except Exception as err:
        return jsonify(success=False, error=str(err)), 500

@app.route('/update_author_form.html', methods=['GET'])
def update_author_form():
    author_id = request.args.get('id')
    firstname = request.args.get('firstname')
    lastname = request.args.get('lastname')
    photo = request.args.get('photo')

    author = {
        'id': author_id,
        'firstname': firstname,
        'lastname': lastname,
        'photo': photo
    }

    return render_template('update_author_form.html', author=author)

@app.route('/author/<int:author_id>', methods=['POST'])
def update_author(author_id):
    try:
        firstname = request.form.get('firstname')
        lastname = request.form.get('lastname')
        photo = request.files.get('photo')

        if photo:
            filename = secure_filename(photo.filename)
            photo_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
            photo.save(photo_path)
            photo_url = f'/uploads/{filename}'  # Adjust this based on your file serving logic
        else:
            photo_url = request.form.get('existing_photo')

        data = {
            'firstname': firstname,
            'lastname': lastname,
            'photo': photo_url
        }

        response = requests.put(f"{API_URL}/authors/{author_id}", json=data)
        if response.status_code != 200:
            return jsonify(success=False, error="Error updating author"), 400

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
                # if photo_path and os.path.exists(os.path.join(app.config['UPLOAD_FOLDER'], filename)):
                #     os.remove(os.path.join(app.config['UPLOAD_FOLDER'], filename))
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

@app.route('/add_book', methods=['GET', 'POST'])
def add_book():
    if request.method == 'POST':
        title = request.form.get('title')
        details = request.form.get('details')  
        author_id = request.form.get('author')
        is_borrowed = request.form.get('is_borrowed', 'off') == 'on'
        photo = request.files['photo']

        if photo:
            filename = secure_filename(photo.filename)
            photo_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
            photo.save(photo_path)
            photo_path = f'photos/{filename}'
        else:
            photo_path = None

        try:
            author_id = int(author_id)
        except ValueError:
            return jsonify(success=False, error="Author ID must be an integer"), 400

        data = {
            'title': title,
            'details': details,  
            'author_id': author_id,
            'is_borrowed': is_borrowed,
            'photo': photo_path
        }

        app.logger.debug(f"Data to send to API: {data}")

        try:
            response = requests.post(f"{API_URL}/books/new", json=data)
            app.logger.debug(f"API Response: {response.status_code}, Content: {response.content}")

            if response.status_code == 200: 
                response_data = response.json()
                app.logger.debug(f"Book added successfully: {response_data}")
                return redirect(url_for("index"))
            else:
                error_message = response.json().get('error', 'Failed to add book')
                app.logger.error(f"Failed to add book: {error_message}")
                return jsonify(success=False, error=error_message), 500

        except Exception as err:
            app.logger.error(f"Failed to add book: {err}")
            return jsonify(success=False, error=str(err)), 500

    try:
        response = requests.get(f"{API_URL}/authors")
        if response.status_code != 200:
            return "Error fetching authors from API", 400
        authors = response.json()
    except Exception as err:
        app.logger.error(f"Failed to fetch authors: {err}")
        return str(err), 500

    return render_template('add_book_form.html', authors=authors)

@app.route('/add_subscriber', methods=['GET', 'POST'])
def add_subscriber():
    if request.method == 'POST':
        firstname = request.form.get('firstname')
        lastname = request.form.get('lastname')
        email = request.form.get('email')

        data = {
            'firstname': firstname,
            'lastname': lastname,
            'email': email
        }
        print(data)

        try:
            response = requests.post(f"{API_URL}/subscribers/new", json=data)
            if response.status_code == 200:
                return redirect(url_for("get_subscribers"))
            else:
                error_message = response.json().get('error', 'Failed to add subscriber')
                app.logger.error(f"Failed to add subscriber: {error_message}")
                return jsonify(success=False, error=error_message), 500

        except Exception as err:
            app.logger.error(f"Failed to add subscriber: {err}")
            return jsonify(success=False, error=str(err)), 500
    
    return render_template('add_subscriber.html')


@app.route('/update_book/<int:book_id>', methods=['GET'])
def update_book_form(book_id):
    try:
        response = requests.get(f"{API_URL}/books/{book_id}")
        if response.status_code != 200:
            return "Error fetching book details from API", 400
        book = response.json()

        response = requests.get(f"{API_URL}/authors")
        if response.status_code != 200:
            return "Error fetching authors from API", 400
        authors = response.json()

        return render_template('update_book_form.html', book=book, authors=authors)
    except Exception as err:
        return str(err), 500

@app.route('/book/<int:book_id>', methods=['POST'])
def update_book(book_id):
    try:
        title = request.form.get('title')
        details = request.form.get('details')
        author_id = request.form.get('author')
        is_borrowed = request.form.get('is_borrowed', 'off') == 'on'
        photo = request.files['photo']

        if photo:
            filename = secure_filename(photo.filename)
            photo_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
            photo.save(photo_path)
            photo_path = f'photos/{filename}'
        else:
            photo_path = request.form.get('existing_photo')

        data = {
            'title': title,
            'details': details,
            'author_id': int(author_id),
            'is_borrowed': is_borrowed,
            'photo': photo_path
        }

        response = requests.put(f"{API_URL}/books/{book_id}", json=data)
        if response.status_code != 200:
            return jsonify(success=False, error="Error updating book"), 400

        return redirect(url_for('book_details', book_id=book_id))
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