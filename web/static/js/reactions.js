document.addEventListener("DOMContentLoaded", () => {
    document.querySelectorAll(".react-btn").forEach(button => {
        button.addEventListener("click", () => {
            const type = button.dataset.type;
            const postId =button.parentElement.dataset.postId;

            fetch(`/post/${postId}/react`, {
                method: "POST",
                headers: {
                    "content-Type": "application/json"
                },
                body: JSON.stringify({ type })
            })
            .then(res => res.json())
            .then(data => {
                console.log("Raction added:", data);
            })
            .catch(err => console.error(err))
        });
    });
});