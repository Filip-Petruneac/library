import React, { useState, useEffect } from "react";
import axios from "axios";

const baseURL = "http://localhost:8080/authors";

function useAuthors() {
    const [data, setData] = useState(null);
    const [error, setError] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchAuthors = async () => {
            try {
                const response = await axios.get(baseURL, { withCredentials: true });
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

    return { data, error, loading };
}

function GetAuthors() {
    const { data, error, loading } = useAuthors();

    if (loading) return <p>Loading...</p>;
    if (error) return <p style={{ color: "red" }}>{error}</p>;

    return (
        <div>
            <h2>JSON Response</h2>
            <pre>{JSON.stringify(data, null, 2)}</pre>
        </div>
    );
}

export default GetAuthors;
