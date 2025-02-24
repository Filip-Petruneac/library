import React, { useState } from "react";
import axios from "axios";
import "../styles/AddAuthor.css";

const baseURL = import.meta.env.VITE_API_URL;

function AddAuthor({ onAddSuccess, onCancel }) {
  const [firstname, setFirstname] = useState("");
  const [lastname, setLastname] = useState("");
  const [photo, setPhoto] = useState(null);
  const [preview, setPreview] = useState(null); // <--- For image preview
  const [error, setError] = useState(null);

  // Handle file input change
  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      setPhoto(e.target.files[0]);
      // Generate a local URL for the file so we can display a preview
      setPreview(URL.createObjectURL(e.target.files[0]));
    }
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Create a FormData object to include file and text data
    const formData = new FormData();
    formData.append("firstname", firstname);
    formData.append("lastname", lastname);
    if (photo) {
      formData.append("photo", photo);
    }

    try {
      const response = await axios.post(`${baseURL}/authors/new`, formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
        withCredentials: true,
      });
      onAddSuccess(response.data);
    } catch (err) {
      console.error("Error adding author:", err);
      setError("Error adding author.");
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-container">
        <button className="modal-close" onClick={onCancel}>
          &times;
        </button>
        <h2 className="modal-title">Add Author</h2>
        <p className="modal-description">
          Fill in the form below to add a new author.
        </p>

        <form onSubmit={handleSubmit} className="modal-form">
          <div>
            <label className="modal-label" htmlFor="firstname">
              First Name:
            </label>
            <input
              id="firstname"
              type="text"
              className="modal-input"
              value={firstname}
              onChange={(e) => setFirstname(e.target.value)}
              required
            />
          </div>

          <div>
            <label className="modal-label" htmlFor="lastname">
              Last Name:
            </label>
            <input
              id="lastname"
              type="text"
              className="modal-input"
              value={lastname}
              onChange={(e) => setLastname(e.target.value)}
              required
            />
          </div>

          <div>
            <label className="modal-label" htmlFor="photo">
              Photo:
            </label>
            <input
              id="photo"
              type="file"
              className="modal-input"
              onChange={handleFileChange}
              accept="image/*"
            />
          </div>

          {preview && (
            <div className="preview-container">
              <img src={preview} alt="Preview" className="preview-image" />
            </div>
          )}

          {error && <p className="error-message">{error}</p>}

          <div className="modal-button-container">
            <button
              type="button"
              className="modal-cancel-button"
              onClick={onCancel}
            >
              Cancel
            </button>
            <button type="submit" className="modal-submit-button">
              Add Author
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default AddAuthor;
