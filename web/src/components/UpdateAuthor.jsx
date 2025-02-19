import React, { useState } from "react";
import axios from "axios";

const baseURL = import.meta.env.VITE_API_URL;

function UpdateAuthor({ author, onUpdateSuccess, onCancel }) {
  const [firstname, setFirstname] = useState(author.firstname);
  const [lastname, setLastname] = useState(author.lastname);
  const [photo, setPhoto] = useState(author.photo);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await axios.put(
        `${baseURL}/authors/${author.id}`,
        { firstname, lastname, photo },
        { withCredentials: true }
      );

      const updatedAuthor = { ...author, firstname, lastname, photo };
      onUpdateSuccess(updatedAuthor);
    } catch (err) {
      setError(err.response?.data?.error || "Failed to update author.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-container">
        <button onClick={onCancel} className="modal-close">
          &times;
        </button>

        <h2 className="modal-title">Update Author</h2>
        <p className="modal-description">
          Modify the author's details below and save your changes.
        </p>

        {error && <p className="error-message">{error}</p>}

        <form onSubmit={handleSubmit} className="modal-form">
          <div>
            <label className="modal-label">Firstname</label>
            <input
              type="text"
              value={firstname}
              onChange={(e) => setFirstname(e.target.value)}
              required
              className="modal-input"
            />
          </div>

          <div>
            <label className="modal-label">Lastname</label>
            <input
              type="text"
              value={lastname}
              onChange={(e) => setLastname(e.target.value)}
              required
              className="modal-input"
            />
          </div>

          <div>
            <label className="modal-label">Photo URL</label>
            <input
              type="text"
              value={photo}
              onChange={(e) => setPhoto(e.target.value)}
              className="modal-input"
            />
          </div>

          <div className="modal-button-container">
            <button type="button" onClick={onCancel} className="modal-cancel-button">
              Cancel
            </button>
            <button type="submit" disabled={loading} className="modal-submit-button">
              {loading ? "Updating..." : "Update"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default UpdateAuthor;
