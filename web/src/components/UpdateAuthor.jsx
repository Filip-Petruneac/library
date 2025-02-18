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
            await axios.put(`${baseURL}/authors/${author.id}`, { firstname, lastname, photo }, { withCredentials: true });

            const UpdatedAuthor = { ...author, firstname, lastname, photo };
            onUpdateSuccess(UpdatedAuthor);
        } catch (err) {
            setError(err.response?.data?.error || "Failed to update author.")
        } finally {
            setLoading(false);
        }
    };

    return (
        <div
          className="
            fixed inset-0 z-50 flex items-center justify-center
            bg-black bg-opacity-50 backdrop-blur-sm
            px-4 py-6
            animate-fadeIn
          "
        >
          {/* Modal Container */}
          <div
            className="
              bg-white rounded-lg shadow-lg relative
              w-full max-w-md
              p-6
              animate-scaleIn
            "
          >
            {/* Close Button */}
            <button
              onClick={onCancel}
              className="
                absolute top-3 right-3
                text-gray-500 hover:text-gray-800
                text-2xl font-bold
                transition-colors
              "
            >
              &times;
            </button>
    
            {/* Title & Description */}
            <h2 className="text-xl font-semibold mb-1">Update Author</h2>
            <p className="text-gray-600 text-sm mb-4">
              Modify the author's details below and save your changes.
            </p>
    
            {/* Error Message */}
            {error && (
              <p className="text-red-500 mb-2">
                {error}
              </p>
            )}
    
            {/* Update Form */}
            <form onSubmit={handleSubmit} className="flex flex-col space-y-4">
              <div>
                <label className="block mb-1 font-medium">Firstname</label>
                <input
                  type="text"
                  value={firstname}
                  onChange={(e) => setFirstname(e.target.value)}
                  required
                  className="
                    border border-gray-300 rounded-md w-full
                    p-2
                    focus:outline-none focus:ring-2 focus:ring-blue-400
                  "
                />
              </div>
    
              <div>
                <label className="block mb-1 font-medium">Lastname</label>
                <input
                  type="text"
                  value={lastname}
                  onChange={(e) => setLastname(e.target.value)}
                  required
                  className="
                    border border-gray-300 rounded-md w-full
                    p-2
                    focus:outline-none focus:ring-2 focus:ring-blue-400
                  "
                />
              </div>
    
              <div>
                <label className="block mb-1 font-medium">Photo URL</label>
                <input
                  type="text"
                  value={photo}
                  onChange={(e) => setPhoto(e.target.value)}
                  className="
                    border border-gray-300 rounded-md w-full
                    p-2
                    focus:outline-none focus:ring-2 focus:ring-blue-400
                  "
                />
              </div>
    
              {/* Action Buttons */}
              <div className="flex justify-end space-x-3 pt-2">
                <button
                  type="button"
                  onClick={onCancel}
                  className="
                    px-4 py-2
                    rounded-md
                    border border-gray-300
                    text-gray-600
                    hover:bg-gray-100
                    transition-colors
                  "
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="
                    px-4 py-2
                    rounded-md
                    bg-blue-500 text-white
                    hover:bg-blue-600
                    transition-colors
                  "
                >
                  {loading ? "Updating..." : "Update"}
                </button>
              </div>
            </form>
          </div>
        </div>
    );
}

export default UpdateAuthor;