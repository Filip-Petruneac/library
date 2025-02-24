import React, { useState, useEffect } from "react";
import axios from "axios";
import DeleteAuthor from "./DeleteAuthor";
import UpdateAuthor from "./UpdateAuthor";
import AddAuthor from "./AddAuthor";
import "../styles/GetAuthors.css";
import "../styles/UpdateAuthor.css";
import "../styles/AddAuthor.css";

const baseURL = import.meta.env.VITE_API_URL;

function useAuthors() {
  const [data, setData] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAuthors = async () => {
      try {
        const response = await axios.get(`${baseURL}/authors`, { withCredentials: true });
        setData(response.data);
      } catch (error) {
        console.error("Error fetching data:", error); // Debugging
        setError(error.response?.data?.error || "An error occurred.");
      } finally {
        setLoading(false);
      }
    };

    fetchAuthors();
  }, []);

  return { data, setData, error, loading };
}

function GetAuthors() {
  const { data, setData, error, loading } = useAuthors();
  const [editingAuthor, setEditingAuthor] = useState(null);
  const [addingAuthor, setAddingAuthor] = useState(false);

  if (loading) return <p>Loading...</p>;
  if (error) return <p style={{ color: "red" }}>{error}</p>;

  const handleUpdateSuccess = (updatedAuthor) => {
    setData((prevData) =>
      prevData.map((author) =>
        author.id === updatedAuthor.id ? updatedAuthor : author
      )
    );
    setEditingAuthor(null);
  };

  const handleDeleteUpdate = (deletedId) => {
    setData((prevData) => prevData.filter((author) => author.id !== deletedId));
  };

  const handleAdd = () => {
    setAddingAuthor(true);
  };

  const handleAddSuccess = (newAuthor) => {
    setData((prevData) => [...prevData, newAuthor]);
    setAddingAuthor(false);
  };

  return (
    <div className="page-container">
      <h1 className="page-title">Authors</h1>

      <button
        onClick={handleAdd}
        className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md mb-4"
      >
        Add Author
      </button>

      {editingAuthor && (
        <UpdateAuthor
          author={editingAuthor}
          onUpdateSuccess={handleUpdateSuccess}
          onCancel={() => setEditingAuthor(null)}
        />
      )}

      {addingAuthor && (
        <AddAuthor
          onAddSuccess={handleAddSuccess}
          onCancel={() => setAddingAuthor(false)}
        />
      )}

      <div className="table-container">
        <table className="authors-table">
          <thead>
            <tr>
              <th>Firstname</th>
              <th>Lastname</th>
              <th>Photo</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {data.map((author) => (
              <tr key={author.id}>
                <td>{author.firstname}</td>
                <td>{author.lastname}</td>
                <td>
                  <img
                    src={author.photo}
                    alt="Author"
                    className="author-image"
                  />
                </td>
                <td>
                  <button
                    onClick={() => setEditingAuthor(author)}
                    className="bg-green-500 hover:bg-green-600 text-white px-2 py-1 rounded-md mr-2"
                  >
                    UPDATE
                  </button>

                  <DeleteAuthor id={author.id} onDeleted={handleDeleteUpdate} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default GetAuthors;
