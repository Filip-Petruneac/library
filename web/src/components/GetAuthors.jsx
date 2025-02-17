import React, { useState, useEffect } from "react";
import axios from "axios";
import DeleteAuthor from "./DeleteAuthor";
import "../GetAuthors.css";

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

    if (loading) return <p>Loading...</p>;
    if (error) return <p style={{ color: "red" }}>{error}</p>;

    const handleUpdate = (id) => console.log("Update ID:", id);
    const handleAdd = (id) => console.log("Add ID:", id);

    // Update the state after a successful deletion.
    const handleDeleteUpdate = (deletedId) => {
        setData((prevData) => prevData.filter((author) => author.id !== deletedId));
    };

    return (
        <div className="page-container">
        <h1 className="page-title">Authors</h1>
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
                    <button onClick={() => handleUpdate(author.id)}>UPDATE</button>
                    <DeleteAuthor id={author.id} onDeleted={handleDeleteUpdate} />
                    <button onClick={() => handleAdd(author.id)}>ADD</button>
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
