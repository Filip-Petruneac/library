import React, { useState } from "react";
import axios from "axios";

const baseURL = import.meta.env.VITE_API_URL

function DeleteAuthor({ id, onDeleted}) {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const handleDelete = async () => {
        if (!window.confirm("Are you sure you want to delete this author?")) return;

        setLoading(true);
        setError(null);

        try {
            await axios.delete(`${baseURL}/authors/${id}`, {withCredentials: true});
            onDeleted(id);
        } catch (err) {
            console.error("Deletion error:", err);
            
            let msg = err.response?.data?.error;
            // 2) If the server returns plain text, it's in err.response?.data
            if (!msg) {
                msg = err.response?.data;
            }
            // 3) If there's still no message, use a fallback
            if (!msg) {
                msg = "An error occurred during deletion.";
            }
            setError(msg);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            {error && <p style={{color: "red"}}>{error}</p>}
            <button onClick={handleDelete} disabled={loading}>
                {loading ? "Deleting..." : "DELETE"}
            </button>
        </div>
    )
};



export default DeleteAuthor;